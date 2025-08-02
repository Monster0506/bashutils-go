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
