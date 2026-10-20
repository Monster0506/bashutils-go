package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/monster0506/bashutils-go/internal/utils"
	"github.com/spf13/cobra"
)

const (
	defaultSplitLines = 1000
	alphaBase         = 26
	numericBase       = 10
)

type splitNamer struct {
	prefix  string
	numeric bool
	width   int
	extra   string
}

func (n splitNamer) name(index int) (string, error) {
	base := alphaBase
	if n.numeric {
		base = numericBase
	}

	capacity := 1
	for range n.width {
		capacity *= base
	}
	if index >= capacity {
		return "", fmt.Errorf("split: output file suffixes exhausted")
	}

	digits := make([]byte, n.width)
	for i := n.width - 1; i >= 0; i-- {
		d := index % base
		index /= base
		if n.numeric {
			digits[i] = byte('0' + d)
		} else {
			digits[i] = byte('a' + d)
		}
	}
	return n.prefix + string(digits) + n.extra, nil
}

var splitCmd = &cobra.Command{
	Use:   "split [file] [prefix]",
	Short: "Split files into pieces",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		linesPerFile, _ := cmd.Flags().GetInt64("lines")
		bytesPerFile, _ := cmd.Flags().GetString("bytes")

		namer, err := newSplitNamer(cmd, args)
		if err != nil {
			return err
		}

		if linesPerFile == 0 && bytesPerFile == "" {
			linesPerFile = defaultSplitLines
		}
		if linesPerFile > 0 && bytesPerFile != "" {
			return fail(1, "split: cannot split by lines and bytes simultaneously")
		}

		inputs := resolveInputs("split", args[:1], reportErrors)
		if len(inputs) == 0 {
			return exitWith(1)
		}

		r := openInput("split", inputs[0], reportErrors)
		if r == nil {
			return exitWith(1)
		}
		defer r.Close()

		if linesPerFile > 0 {
			err = splitByLines(r, namer, linesPerFile)
		} else {
			err = splitByBytes(r, namer, bytesPerFile)
		}
		if err != nil {
			return fail(1, "%v", err)
		}
		return nil
	},
}

func init() {
	splitCmd.Flags().Int64P("lines", "l", 0, "split by number of lines")
	splitCmd.Flags().StringP("bytes", "b", "", "split by number of bytes (e.g., '1K', '1M')")
	splitCmd.Flags().BoolP("numeric-suffixes", "d", false, "use numeric suffixes instead of alphabetic")
	splitCmd.Flags().IntP("suffix-length", "a", 2, "generate suffixes of length N")
	splitCmd.Flags().String("additional-suffix", "", "append an additional SUFFIX to file names")
}

func newSplitNamer(cmd *cobra.Command, args []string) (splitNamer, error) {
	width, _ := cmd.Flags().GetInt("suffix-length")
	if width < 1 {
		return splitNamer{}, fail(1, "split: invalid suffix length: %d", width)
	}

	n := splitNamer{prefix: "x", width: width}
	if len(args) > 1 {
		n.prefix = args[1]
	}
	n.numeric, _ = cmd.Flags().GetBool("numeric-suffixes")
	n.extra, _ = cmd.Flags().GetString("additional-suffix")
	return n, nil
}

type pieceWriter struct {
	file *os.File
	buf  *bufio.Writer
}

func createPiece(namer splitNamer, index int) (*pieceWriter, error) {
	name, err := namer.name(index)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("split: creating output file: %v", err)
	}
	return &pieceWriter{file: f, buf: bufio.NewWriter(f)}, nil
}

func (p *pieceWriter) close() error {
	err := p.buf.Flush()
	if cerr := p.file.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("split: writing output file: %v", err)
	}
	return nil
}

func splitByLines(r io.Reader, namer splitNamer, perFile int64) error {
	scanner := utils.NewScanner(r)
	var piece *pieceWriter
	index, inPiece := 0, int64(0)

	for scanner.Scan() {
		if piece == nil {
			p, err := createPiece(namer, index)
			if err != nil {
				return err
			}
			piece, index, inPiece = p, index+1, 0
		}

		piece.buf.WriteString(scanner.Text())
		if !scanner.Unterminated() {
			piece.buf.WriteByte('\n')
		}

		inPiece++
		if inPiece < perFile {
			continue
		}
		if err := piece.close(); err != nil {
			return err
		}
		piece = nil
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("split: reading input: %v", err)
	}
	if piece != nil {
		return piece.close()
	}
	return nil
}

func parseBytesString(bytesStr string) (int64, error) {
	bytesStr = strings.TrimSpace(bytesStr)
	lastChar := ' '
	if len(bytesStr) > 0 {
		lastChar = rune(bytesStr[len(bytesStr)-1])
	}

	multiplier := int64(1)
	valueStr := bytesStr

	switch lastChar {
	case 'k', 'K':
		multiplier = 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'm', 'M':
		multiplier = 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	case 'g', 'G':
		multiplier = 1024 * 1024 * 1024
		valueStr = bytesStr[:len(bytesStr)-1]
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size format: %s", bytesStr)
	}
	if value <= 0 {
		return 0, fmt.Errorf("invalid byte size: %s", bytesStr)
	}
	return value * multiplier, nil
}

func splitByBytes(r io.Reader, namer splitNamer, sizeStr string) error {
	perFile, err := parseBytesString(sizeStr)
	if err != nil {
		return fmt.Errorf("split: %v", err)
	}

	reader := bufio.NewReader(r)
	for index := 0; ; index++ {
		if _, err := reader.Peek(1); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("split: reading input: %v", err)
		}

		piece, err := createPiece(namer, index)
		if err != nil {
			return err
		}

		_, copyErr := io.CopyN(piece.buf, reader, perFile)
		if err := piece.close(); err != nil {
			return err
		}

		if copyErr == io.EOF {
			return nil
		}
		if copyErr != nil {
			return fmt.Errorf("split: copying to output file: %v", copyErr)
		}
	}
}
