package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bashutils",
	Short: "A Go-based reimplementation of bash coreutils",
	Long:  `bashutils is a CLI tool written in Go that mimics common bash coreutils commands.`,
	Args:  cobra.MinimumNArgs(1),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(echoCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(headCmd)
	rootCmd.AddCommand(tailCmd)
	rootCmd.AddCommand(wcCmd)
	rootCmd.AddCommand(cutCmd)
	rootCmd.AddCommand(sortCmd)
	rootCmd.AddCommand(uniqCmd)
	rootCmd.AddCommand(grepCmd)
	rootCmd.AddCommand(trCmd)
	rootCmd.AddCommand(pasteCmd)
	rootCmd.AddCommand(splitCmd)
	rootCmd.AddCommand(xargsCmd)
}
