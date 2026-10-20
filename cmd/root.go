package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bashutils",
	Short: "A Go-based reimplementation of bash coreutils",
	Long:  `bashutils is a CLI tool written in Go that mimics common bash coreutils commands.`,
	Args:  cobra.MinimumNArgs(1),
}

func Execute() {
	dispatchByName()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func dispatchByName() {
	base := filepath.Base(os.Args[0])
	name := strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))

	for _, c := range rootCmd.Commands() {
		if c.Name() != name {
			continue
		}

		rootCmd.SetArgs(append([]string{name}, os.Args[1:]...))
		return
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
	rootCmd.AddCommand(calCmd)
	rootCmd.AddCommand(installCmd)
}
