package utils

import (
	"bufio"
	"bytes"
	"io"
	"math"
)

type LineScanner struct {
	*bufio.Scanner
	unterminated bool
}

func NewScanner(r io.Reader) *LineScanner {
	s := &LineScanner{Scanner: bufio.NewScanner(r)}
	s.Buffer(nil, math.MaxInt)
	s.Split(s.splitLines)
	return s
}

func (s *LineScanner) Unterminated() bool {
	return s.unterminated
}

func (s *LineScanner) splitLines(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		s.unterminated = false
		return i + 1, data[:i], nil
	}
	if atEOF {
		s.unterminated = true
		return len(data), data, nil
	}
	return 0, nil, nil
}

func ReadLines(r io.Reader) ([]string, error) {
	scanner := NewScanner(r)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
