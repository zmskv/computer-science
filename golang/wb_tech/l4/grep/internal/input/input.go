package input

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"mygrep/internal/search"
)

const maxScanTokenSize = 1024 * 1024

func ReadAll(stdin io.Reader, files []string) ([]search.Line, error) {
	if len(files) == 0 {
		return readFromReader(stdin, "")
	}

	var all []search.Line
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}

		lines, readErr := readFromReader(f, path)
		closeErr := f.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close %s: %w", path, closeErr)
		}

		all = append(all, lines...)
	}

	return all, nil
}

func readFromReader(r io.Reader, source string) ([]search.Line, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxScanTokenSize)

	lines := make([]search.Line, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		lines = append(lines, search.Line{
			Source: source,
			Number: lineNo,
			Text:   scanner.Text(),
		})
	}

	if err := scanner.Err(); err != nil {
		if source == "" {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return nil, fmt.Errorf("read %s: %w", source, err)
	}

	return lines, nil
}
