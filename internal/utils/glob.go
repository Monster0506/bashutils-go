package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExpandGlobs takes a slice of arguments and expands any glob patterns
// into matching file paths. It returns a new slice with globs expanded.
func ExpandGlobs(args []string) ([]string, error) {
	var expanded []string
	
	for _, arg := range args {
		// Check if the argument contains any glob characters
		if containsGlobChars(arg) {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return nil, fmt.Errorf("glob error for pattern %s: %v", arg, err)
			}
			
			// If no matches found, keep the original pattern (bash behavior)
			if len(matches) == 0 {
				expanded = append(expanded, arg)
				continue
			}
			
			// Sort matches for consistent output
			sort.Strings(matches)
			expanded = append(expanded, matches...)
		} else {
			// No glob characters, add as-is
			expanded = append(expanded, arg)
		}
	}
	
	return expanded, nil
}

// containsGlobChars checks if a string contains any glob pattern characters
func containsGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ExpandGlobsWithValidation expands globs and validates that files exist
// It's similar to ExpandGlobs but filters out non-existent files
func ExpandGlobsWithValidation(args []string) ([]string, error) {
	expanded, err := ExpandGlobs(args)
	if err != nil {
		return nil, err
	}
	
	var valid []string
	for _, path := range expanded {
		if _, err := os.Stat(path); err == nil {
			valid = append(valid, path)
		} else if os.IsNotExist(err) {
			// Skip non-existent files
			continue
		} else {
			// Return error for other issues (permission denied, etc.)
			return nil, fmt.Errorf("error accessing %s: %v", path, err)
		}
	}
	
	return valid, nil
}

// ExpandGlobsForReading expands globs and returns only readable files
func ExpandGlobsForReading(args []string) ([]string, error) {
	expanded, err := ExpandGlobs(args)
	if err != nil {
		return nil, err
	}
	
	var readable []string
	for _, path := range expanded {
		if file, err := os.Open(path); err == nil {
			file.Close()
			readable = append(readable, path)
		} else if os.IsNotExist(err) {
			// Skip non-existent files
			continue
		} else {
			// Return error for other issues (permission denied, etc.)
			return nil, fmt.Errorf("error accessing %s: %v", path, err)
		}
	}
	
	return readable, nil
} 