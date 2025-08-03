package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadAllFromFilesOrStdin(files []string) (string, error) {
	if len(files) == 0 {
		return ReadAllFromReader(os.Stdin)
	}
	
	// Expand glob patterns in file arguments
	expandedFiles, err := ExpandGlobsForReading(files)
	if err != nil {
		return "", err
	}
	
	var sb strings.Builder
	for _, f := range expandedFiles {
		file, err := os.Open(f)
		if err != nil {
			return "", fmt.Errorf("%s: %v", f, err)
		}
		defer file.Close()
		content, err := ReadAllFromReader(file)
		if err != nil {
			return "", err
		}
		sb.WriteString(content)
	}
	return sb.String(), nil
}

func ReadAllFromReader(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ReadLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ReadLinesFromFilesOrStdin reads lines from files or stdin if no files provided
// This is the common pattern used by commands like sort, uniq, grep, etc.
func ReadLinesFromFilesOrStdin(files []string) ([]string, error) {
	if len(files) == 0 {
		// Read from stdin when no files provided
		return ReadLines(os.Stdin)
	}
	
	// Expand glob patterns in file arguments
	expandedFiles, err := ExpandGlobsForReading(files)
	if err != nil {
		return nil, err
	}
	
	var allLines []string
	for _, path := range expandedFiles {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", path, err)
		}
		
		lines, err := ReadLines(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %v", path, err)
		}
		
		allLines = append(allLines, lines...)
	}
	
	return allLines, nil
}
