package cmd

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	grepIncludes    []string
	grepExcludes    []string
	grepExcludeDirs []string
)

func grepInputs(cmd *cobra.Command, args []string, mode errorMode) ([]input, bool) {
	recursive, _ := cmd.Flags().GetBool("recursive")
	dereference, _ := cmd.Flags().GetBool("dereference-recursive")
	if !recursive && !dereference {
		return resolveInputs("grep", args, mode), false
	}

	implicit := len(args) == 0
	if implicit {
		args = []string{"."}
	}

	var inputs []input
	sawDir := implicit
	for _, root := range resolveInputs("grep", args, mode) {
		if root.name == stdinName {
			inputs = append(inputs, root)
			continue
		}

		files, isDir := collectFiles(root.name, mode)
		sawDir = sawDir || isDir
		for _, f := range files {
			inputs = append(inputs, input{name: f, label: f})
		}
	}

	for i := range inputs {
		inputs[i].index = i
		inputs[i].count = len(inputs)
	}
	return inputs, sawDir
}

func collectFiles(root string, mode errorMode) ([]string, bool) {
	info, err := os.Stat(root)
	if err != nil {
		inputError(mode, "grep: %v\n", err)
		return nil, false
	}

	if !info.IsDir() {
		if fileAllowed(filepath.Base(root)) {
			return []string{root}, false
		}
		return nil, false
	}

	var files []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && matchesAny(grepExcludeDirs, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if fileAllowed(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, true
}

func fileAllowed(name string) bool {
	if len(grepIncludes) > 0 && !matchesAny(grepIncludes, name) {
		return false
	}
	return !matchesAny(grepExcludes, name)
}

func matchesAny(globs []string, name string) bool {
	for _, g := range globs {
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
	}
	return false
}
