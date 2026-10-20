package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// ExpandEnvironmentVariables expands both Unix-style ($VAR) and Windows-style (%VAR%) environment variables
func ExpandEnvironmentVariables(input string) string {
	// First expand Unix-style variables ($VAR)
	unixRegex := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	result := unixRegex.ReplaceAllStringFunc(input, func(match string) string {
		varName := match[1:] // Remove the $ prefix
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if not found
	})

	// Then expand Windows-style variables (%VAR%)
	winRegex := regexp.MustCompile(`%([A-Za-z_][A-Za-z0-9_]*)%`)
	result = winRegex.ReplaceAllStringFunc(result, func(match string) string {
		varName := match[1 : len(match)-1] // Remove the % prefix and suffix
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if not found
	})

	return result
}
