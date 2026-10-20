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
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		mode := keepExisting
		if force {
			mode = replaceExisting
		}

		exe, err := os.Executable()
		if err != nil {
			return fail(1, "install: %v", err)
		}

		dir := args[0]
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return fail(1, "install: %v", err)
		}

		ext := filepath.Ext(exe)
		for _, c := range rootCmd.Commands() {
			if nonToolCmds[c.Name()] {
				continue
			}

			dst := filepath.Join(dir, c.Name()+ext)
			if err := installOne(exe, dst, mode); err != nil {
				warn("install: %v\n", err)
				continue
			}

			fmt.Println(dst)
		}

		return nil
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
