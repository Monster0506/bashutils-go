package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

type countSpec struct {
	n         int
	fromStart bool
}

func parseCountSpec(s string) (countSpec, error) {
	var spec countSpec
	if strings.HasPrefix(s, "+") {
		spec.fromStart = true
	}

	n, err := strconv.Atoi(strings.TrimLeft(s, "+-"))
	if err != nil || n < 0 {
		return spec, fmt.Errorf("invalid number of lines or bytes: %q", s)
	}
	spec.n = n
	return spec, nil
}

var tailCmd = &cobra.Command{
	Use:   "tail [files...]",
	Short: "Output the last part of files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flag := "lines"
		byBytes := cmd.Flags().Changed("bytes")
		if byBytes {
			flag = "bytes"
		}

		value, _ := cmd.Flags().GetString(flag)
		spec, err := parseCountSpec(value)
		if err != nil {
			return fail(1, "tail: %v", err)
		}
		headers := headerModeOf(cmd)

		eachInput("tail", args, func(in input, r io.Reader) {
			printHeader(in, headers)

			var err error
			if byBytes {
				err = tailBytes(r, spec)
			} else {
				err = tailLines(r, spec)
			}
			if err != nil {
				warn("tail: %s: %v\n", in.name, err)
			}
		})
		return nil
	},
}

func init() {
	tailCmd.Flags().StringP("lines", "n", "10", "output the last NUM lines; with +NUM, start at line NUM")
	tailCmd.Flags().StringP("bytes", "c", "", "output the last NUM bytes; with +NUM, start at byte NUM")
	tailCmd.Flags().BoolP("quiet", "q", false, "never print file name headers")
	tailCmd.Flags().BoolP("verbose", "v", false, "always print file name headers")
}

func tailLines(r io.Reader, spec countSpec) error {
	scanner := utils.NewScanner(r)
	if spec.fromStart {
		for i := 1; scanner.Scan(); i++ {
			if i >= spec.n {
				printLine(scanner)
			}
		}
		return scanner.Err()
	}

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > spec.n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for i, line := range lines {
		if i == len(lines)-1 && scanner.Unterminated() {
			fmt.Fprint(stdout, line)
			continue
		}
		fmt.Fprintln(stdout, line)
	}
	return nil
}

func tailBytes(r io.Reader, spec countSpec) error {
	if spec.fromStart {
		if spec.n > 1 {
			if _, err := io.CopyN(io.Discard, r, int64(spec.n-1)); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
		_, err := io.Copy(stdout, r)
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if spec.n < len(data) {
		data = data[len(data)-spec.n:]
	}
	_, err = stdout.Write(data)
	return err
}
