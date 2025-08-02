package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat [file]",
	Short: "Concatenate and display files",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Expand glob patterns in arguments
		expandedArgs, err := utils.ExpandGlobsForReading(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			return
		}
		
		for _, path := range expandedArgs {
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
