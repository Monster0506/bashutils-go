package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat [file]",
	Short: "Concatenate and display files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cat: %v\n", err)
				continue
			}
			func() (n int, err error) {
				if sw, ok := io.Writer(os.Stdout).(io.StringWriter); ok {
					return sw.WriteString(string(data))
				}
				return io.Writer(os.Stdout).Write([]byte(string(data)))
			}()
		}
	},
}
