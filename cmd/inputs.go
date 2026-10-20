package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/monster0506/bashutils-go/internal/utils"
)

const stdinName = "standard input"

type input struct {
	name  string
	index int
	count int
}

func eachInput(cmdName string, args []string, fn func(in input, r io.Reader)) {
	if len(args) == 0 {
		args = []string{"-"}
	}

	paths, err := utils.ExpandGlobs(args)
	if err != nil {
		warn("%s: %v\n", cmdName, err)
		return
	}

	for i, path := range paths {
		in := input{name: path, index: i, count: len(paths)}
		if path == "-" {
			in.name = stdinName
			fn(in, os.Stdin)
			continue
		}

		openAndRun(cmdName, in, fn)
	}
}

func openAndRun(cmdName string, in input, fn func(in input, r io.Reader)) {
	f, err := os.Open(in.name)
	if err != nil {
		warn("%s: %v\n", cmdName, err)
		return
	}
	defer f.Close()

	fn(in, f)
}

func printHeader(in input) {
	if in.count < 2 {
		return
	}

	if in.index > 0 {
		fmt.Println()
	}
	fmt.Printf("==> %s <==\n", in.name)
}

func printLine(s *utils.LineScanner) {
	fmt.Print(s.Text())
	if !s.Unterminated() {
		fmt.Println()
	}
}
