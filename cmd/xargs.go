package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var xargsCmd = &cobra.Command{
	Use:   "xargs [command] [args...]",
	Short: "Build and execute command lines from standard input",
	Long: `xargs reads items from standard input, delimited by blanks (which can be protected
with double or single quotes or a backslash) or newlines, and executes the command
(default is /bin/echo) one or more times with any initial-arguments followed by
items read from standard input.`,
	Args: cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags manually since we disabled flag parsing
		maxArgs := 0
		replaceStr := ""
		nullTerminated := false
		delimiter := ""
		noRunIfEmpty := false
		verbose := false
		
		// Parse xargs flags from the beginning of args
		var commandArgs []string
		commandFound := false
		
		// First pass: check for verbose flag
		for i := 0; i < len(args); i++ {
			if args[i] == "-t" || args[i] == "--verbose" {
				verbose = true
				break
			}
		}
		
		if verbose {
			fmt.Fprintf(os.Stderr, "Parsing args: %v\n", args)
		}
		for i := 0; i < len(args) && !commandFound; i++ {
			arg := args[i]
			switch arg {
			case "-n", "--max-args":
				if i+1 < len(args) {
					maxArgs = parseInt(args[i+1])
					i++ // skip the value
				}
			case "-I", "--replace":
				if i+1 < len(args) {
					replaceStr = args[i+1]
					i++ // skip the value
				}
			case "-0", "--null":
				nullTerminated = true
			case "-d", "--delimiter":
				if i+1 < len(args) {
					delimiter = args[i+1]
					i++ // skip the value
				}
			case "-r", "--no-run-if-empty":
				noRunIfEmpty = true
			case "-t", "--verbose":
				verbose = true
			case "-h", "--help":
				cmd.Help()
				return
			default:
				// If it starts with - but isn't a recognized flag, it might be part of the command
				if strings.HasPrefix(arg, "-") {
					// Check if it's a short flag that might be part of the command
					if len(arg) == 2 && arg[1] != '-' {
						// This could be a command flag, so treat everything from here as command args
						commandArgs = args[i:]
						break
					}
				}
				// This is the start of the command - take everything from here
				commandArgs = args[i:]
				commandFound = true
				if verbose {
					fmt.Fprintf(os.Stderr, "Command args: %v\n", commandArgs)
				}
				break
			}
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "After parsing, commandArgs: %v\n", commandArgs)
		}

		// Read items from stdin
		items, err := readItemsFromStdin(nullTerminated, delimiter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "xargs: error reading input: %v\n", err)
			os.Exit(1)
		}

		if len(items) == 0 {
			if noRunIfEmpty {
				return
			}
			// If no items and no-run-if-empty is false, still run command once with no args
			if len(commandArgs) > 0 {
				executeCommand(commandArgs, nil, replaceStr, verbose)
			}
			return
		}

		// If no command specified, use bashutils echo
		if len(commandArgs) == 0 {
			commandArgs = []string{"bashutils", "echo"}
		}

		// Execute commands
		if verbose {
			fmt.Fprintf(os.Stderr, "About to execute with commandArgs: %v\n", commandArgs)
		}
		if maxArgs > 0 {
			// Split items into chunks of maxArgs
			for i := 0; i < len(items); i += maxArgs {
				end := i + maxArgs
				if end > len(items) {
					end = len(items)
				}
				chunk := items[i:end]
				executeCommand(commandArgs, chunk, replaceStr, verbose)
			}
		} else {
			// Execute all items at once
			executeCommand(commandArgs, items, replaceStr, verbose)
		}
	},
}

func readItemsFromStdin(nullTerminated bool, delimiter string) ([]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var items []string

	if nullTerminated {
		// Read null-terminated items
		scanner.Split(bufio.ScanBytes)
		var currentItem strings.Builder
		for scanner.Scan() {
			if scanner.Text() == "\x00" {
				items = append(items, strings.TrimSpace(currentItem.String()))
				currentItem.Reset()
			} else {
				currentItem.WriteString(scanner.Text())
			}
		}
		// Add the last item if it doesn't end with null
		if currentItem.Len() > 0 {
			items = append(items, strings.TrimSpace(currentItem.String()))
		}
	} else if delimiter != "" {
		// Read items separated by custom delimiter
		scanner.Split(bufio.ScanBytes)
		var currentItem strings.Builder
		for scanner.Scan() {
			if scanner.Text() == delimiter {
				items = append(items, strings.TrimSpace(currentItem.String()))
				currentItem.Reset()
			} else {
				currentItem.WriteString(scanner.Text())
			}
		}
		// Add the last item
		if currentItem.Len() > 0 {
			items = append(items, strings.TrimSpace(currentItem.String()))
		}
	} else {
		// Read items separated by whitespace/newlines
		for scanner.Scan() {
			line := scanner.Text()
			// Split by whitespace and add non-empty items
			fields := strings.Fields(line)
			for _, field := range fields {
				if field != "" {
					items = append(items, field)
				}
			}
		}
	}

	return items, scanner.Err()
}

func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 0
}

func executeCommand(command []string, args []string, replaceStr string, verbose bool) {
	if replaceStr != "" {
		// Replace the replace string with the arguments
		for _, arg := range args {
			replaced := strings.ReplaceAll(strings.Join(command, " "), replaceStr, arg)
			cmdParts := strings.Fields(replaced)
			if len(cmdParts) > 0 {
				executeSingleCommand(cmdParts, verbose)
			}
		}
	} else {
		// Build the final command arguments
		var finalArgs []string
		finalArgs = append(finalArgs, command...)
		finalArgs = append(finalArgs, args...)
		if verbose {
			fmt.Fprintf(os.Stderr, "Final command: %v\n", finalArgs)
		}
		executeSingleCommand(finalArgs, verbose)
	}
}

func executeSingleCommand(args []string, verbose bool) {
	if verbose {
		fmt.Fprintf(os.Stderr, "Executing: %s\n", strings.Join(args, " "))
	}

	// Check if the first argument is "bashutils" and handle it specially
	if len(args) > 0 && args[0] == "bashutils" {
		// Execute bashutils subcommand directly
		if len(args) > 1 {
			// Create a new command with the subcommand
			subCmd := exec.Command(os.Args[0], args[1:]...)
			subCmd.Stdout = os.Stdout
			subCmd.Stderr = os.Stderr
			subCmd.Stdin = os.Stdin

			err := subCmd.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "xargs: command failed: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Regular command execution
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "xargs: command failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	xargsCmd.Flags().IntP("max-args", "n", 0, "use at most max-args arguments per command line")
	xargsCmd.Flags().StringP("replace", "I", "", "replace occurrences of replace-str in the initial-arguments with names read from standard input")
	xargsCmd.Flags().BoolP("null", "0", false, "input items are terminated by a null character instead of by whitespace")
	xargsCmd.Flags().StringP("delimiter", "d", "", "input items are terminated by the specified character")
	xargsCmd.Flags().BoolP("no-run-if-empty", "r", false, "if the standard input does not contain any nonblanks, do not run the command")
	xargsCmd.Flags().BoolP("verbose", "t", false, "print the command line on the standard error output before executing it")
}