package cmd

import (
	"bufio"
	"os"
	"sync"
)

const (
	outputBufferSize = 64 * 1024
	unbufferedSize   = 1
)

var (
	stdout = newStdout()
	outMu  sync.Mutex
)

func newStdout() *bufio.Writer {
	if isTerminal(os.Stdout) {
		return bufio.NewWriterSize(os.Stdout, unbufferedSize)
	}
	return bufio.NewWriterSize(os.Stdout, outputBufferSize)
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func flushOut() {
	outMu.Lock()
	defer outMu.Unlock()

	stdout.Flush()
}
