package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var headCmd = &cobra.Command{
	Use:   "head [files...]",
	Short: "Output the first part of files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		count, _ := cmd.Flags().GetInt("bytes")
		byBytes := cmd.Flags().Changed("bytes")
		headers := headerModeOf(cmd)

		eachInput("head", args, func(in input, r io.Reader) {
			printHeader(in, headers)

			var err error
			if byBytes {
				err = headBytes(r, count)
			} else {
				err = headLines(r, lines)
			}
			if err != nil {
				warn("head: %s: %v\n", in.name, err)
			}
		})
		return nil
	},
}

func init() {
	headCmd.Flags().IntP("lines", "n", 10, "print the first NUM lines; with a negative NUM, all but the last -NUM")
	headCmd.Flags().IntP("bytes", "c", 0, "print the first NUM bytes; with a negative NUM, all but the last -NUM")
	headCmd.Flags().BoolP("quiet", "q", false, "never print file name headers")
	headCmd.Flags().BoolP("verbose", "v", false, "always print file name headers")
}

func headLines(r io.Reader, n int) error {
	scanner := utils.NewScanner(r)
	if n >= 0 {
		for i := 0; i < n && scanner.Scan(); i++ {
			printLine(scanner)
		}
		return scanner.Err()
	}

	var pending []string
	for scanner.Scan() {
		pending = append(pending, scanner.Text())
		if len(pending) > -n {
			fmt.Println(pending[0])
			pending = pending[1:]
		}
	}
	return scanner.Err()
}

func headBytes(r io.Reader, n int) error {
	if n >= 0 {
		_, err := io.CopyN(os.Stdout, r, int64(n))
		if err == io.EOF {
			return nil
		}
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if keep := len(data) + n; keep > 0 {
		_, err = os.Stdout.Write(data[:keep])
	}
	return err
}
