package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

type overwrite int

const (
	keepExisting overwrite = iota
	replaceExisting
)

const dirPerm = 0o755

var nonToolCmds = map[string]bool{
	"help":       true,
	"completion": true,
	"install":    true,
}

var installCmd = &cobra.Command{
	Use:   "install <dir>",
	Short: "Create one executable per command in <dir>",
	Long: `Create one executable per command in <dir>
Rerun with --force after rebuilding to refresh them.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		mode := keepExisting
		if force {
			mode = replaceExisting
		}

		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "install: %v\n", err)
			os.Exit(1)
		}

		dir := args[0]
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			fmt.Fprintf(os.Stderr, "install: %v\n", err)
			os.Exit(1)
		}

		failed := false
		ext := filepath.Ext(exe)
		for _, c := range rootCmd.Commands() {
			if nonToolCmds[c.Name()] {
				continue
			}

			dst := filepath.Join(dir, c.Name()+ext)
			if err := installOne(exe, dst, mode); err != nil {
				fmt.Fprintf(os.Stderr, "install: %v\n", err)
				failed = true
				continue
			}

			fmt.Println(dst)
		}

		if failed {
			os.Exit(1)
		}
	},
}

func init() {
	installCmd.Flags().BoolP("force", "f", false, "replace existing files")
}

func installOne(exe, dst string, mode overwrite) error {
	if _, err := os.Lstat(dst); err != nil {
		return utils.LinkOrCopy(exe, dst)
	}

	if mode == keepExisting {
		return fmt.Errorf("%s exists (use --force)", dst)
	}

	if err := os.Remove(dst); err != nil {
		return err
	}

	return utils.LinkOrCopy(exe, dst)
}
