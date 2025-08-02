package cmd

import (
	"bufio"
	"fmt"
	"github.com/monster0506/bashutils-go/internal/utils"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split [file] [prefix]",
	Short: "Split files into pieces",
	Args:  cobra.RangeArgs(1, 2), // file and optional prefix
	Run: func(cmd *cobra.Command, args []string) {
		linesPerFile, _ := cmd.Flags().GetInt64("lines")
		bytesPerFile, _ := cmd.Flags().GetString("bytes")
		numericSuffixes, _ := cmd.Flags().GetBool("numeric-suffixes")

		filePath := args[0]
		prefix := "x" // Default prefix
		if len(args) > 1 {
			prefix = args[1]
		}

		if linesPerFile == 0 && bytesPerFile == "" {
			linesPerFile = 1000 // Default to 1000 lines if no flag specified
		}
		if linesPerFile > 0 && bytesPerFile != "" {
			fmt.Fprintf(os.Stderr, "split: cannot split by lines and bytes simultaneously\n")
			return
		}

		// Expand glob patterns in file argument
		expandedFiles, err := utils.ExpandGlobsForReading([]string{filePath})
		if err != nil {
			fmt.Fprintf(os.Stderr, "split: %v\n", err)
			return
		}

		// For split, we only process the first file if multiple files match
		if len(expandedFiles) == 0 {
			fmt.Fprintf(os.Stderr, "split: no matching files found\n")
			return
		}

		inputFile, err := os.Open(expandedFiles[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "split: %v\n", err)
			return
		}
		defer inputFile.Close()

		if linesPerFile > 0 {
			splitByLines(inputFile, prefix, linesPerFile, numericSuffixes)
		} else if bytesPerFile != "" {
			splitByBytes(inputFile, prefix, bytesPerFile, numericSuffixes)
		}
	},
}

func init() {
	splitCmd.Flags().Int64P("lines", "l", 0, "split by number of lines")
	splitCmd.Flags().StringP("bytes", "b", "", "split by number of bytes (e.g., '1K', '1M')")
	splitCmd.Flags().BoolP("numeric-suffixes", "d", false, "use numeric suffixes instead of alphabetic")
}

func generateSuffix(index int, numeric bool) string {
	if numeric {
		return fmt.Sprintf("%02d", index) // Pad with leading zeros for consistency
	}
	// Mimic classic 'aa', 'ab', 'ac'... suffixes
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	suf := ""
	for {
		suf = string(alphabet[index%len(alphabet)]) + suf
		index = index / len(alphabet)
		if index == 0 {
			break
		}
		index-- // Adjust for 0-based indexing after division
	}
	return suf
}

func splitByLines(inputFile *os.File, prefix string, linesPerFile int64, numericSuffixes bool) {
	scanner := bufio.NewScanner(inputFile)
	fileIndex := 0
	currentLineCount := int64(0)
	var outputFile *os.File
	var outputWriter *bufio.Writer
	var err error

	for scanner.Scan() {
		if currentLineCount == 0 {
			if outputFile != nil {
				outputWriter.Flush()
				outputFile.Close()
			}
			suffix := generateSuffix(fileIndex, numericSuffixes)
			outputFileName := filepath.Join(filepath.Dir(inputFile.Name()), prefix+suffix)
			outputFile, err = os.Create(outputFileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: creating output file: %v\n", err)
				return
			}
			outputWriter = bufio.NewWriter(outputFile)
			fileIndex++
		}

		_, err = outputWriter.WriteString(scanner.Text() + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "split: writing to output file: %v\n", err)
			return
		}
		currentLineCount++

		if currentLineCount >= linesPerFile {
			currentLineCount = 0 // Reset for the next file
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "split: reading input: %v\n", err)
	}
	if outputFile != nil {
		outputWriter.Flush()
		outputFile.Close()
	}
}

func parseBytesString(bytesStr string) (int64, error) {
	bytesStr = strings.TrimSpace(bytesStr)
	lastChar := ' '
	if len(bytesStr) > 0 {
		lastChar = rune(bytesStr[len(bytesStr)-1])
	}

	multiplier := int64(1)
	valueStr := bytesStr

	switch lastChar {
	case 'k', 'K':
		multiplier = 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'm', 'M':
		multiplier = 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'g', 'G':
		multiplier = 1024 * 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size format: %s", bytesStr)
	}
	return value * multiplier, nil
}

func splitByBytes(inputFile *os.File, prefix string, bytesPerFileStr string, numericSuffixes bool) {
	bytesPerFile, err := parseBytesString(bytesPerFileStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "split: %v\n", err)
		return
	}

	fileIndex := 0
	buffer := make([]byte, 4096) // Use a common buffer size
	var outputFile *os.File
	var errClose error

	for {
		if outputFile == nil {
			suffix := generateSuffix(fileIndex, numericSuffixes)
			outputFileName := filepath.Join(filepath.Dir(inputFile.Name()), prefix+suffix)
			outputFile, err = os.Create(outputFileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: creating output file: %v\n", err)
				return
			}
			fileIndex++
		}

		bytesReadThisFile := int64(0)
		for bytesReadThisFile < bytesPerFile {
			bytesToRead := min(int64(len(buffer)), bytesPerFile-bytesReadThisFile)

			n, err := inputFile.Read(buffer[:bytesToRead])
			if n > 0 {
				_, writeErr := outputFile.Write(buffer[:n])
				if writeErr != nil {
					fmt.Fprintf(os.Stderr, "split: writing to output file: %v\n", writeErr)
					return
				}
				bytesReadThisFile += int64(n)
			}
			if err == io.EOF {
				break // End of input file
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "split: reading input: %v\n", err)
				return
			}
		}

		if outputFile != nil {
			errClose = outputFile.Close()
			if errClose != nil {
				fmt.Fprintf(os.Stderr, "split: closing output file: %v\n", errClose)
				return
			}
			outputFile = nil // Reset for next file
		}

		_, err = inputFile.Seek(0, io.SeekCurrent) // Check if EOF reached after closing and seeking
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "split: seeking input: %v\n", err)
			return
		}
		// If bytesReadThisFile < bytesPerFile and it's not EOF, it means we've read all available data.
		if bytesReadThisFile == 0 && err == io.EOF {
			break // No more data to read from the input file
		}
	}
}
