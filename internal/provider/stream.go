// Package provider implements I/O utilities for reading and writing data streams.
package provider

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

const scannerInitBufSize = 64 * 1024

// ScannerLineReader wraps bufio.Scanner to expose a simple line-by-line Next() API.
type ScannerLineReader struct {
	scanner *bufio.Scanner
}

// NewLineReader creates a ScannerLineReader that reads lines from r, capping each line at maxLineSize bytes.
func NewLineReader(r io.Reader, maxLineSize int) *ScannerLineReader {
	scanner := bufio.NewScanner(r)

	buf := make([]byte, 0, scannerInitBufSize)
	scanner.Buffer(buf, maxLineSize)

	scanner.Split(bufio.ScanLines)

	return &ScannerLineReader{
		scanner: scanner,
	}
}

// Next returns the next line from the reader, or io.EOF when the input is exhausted.
func (r *ScannerLineReader) Next() (string, error) {
	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}

	if err := r.scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %w", err)
	}

	return "", io.EOF
}

// WriteLines writes a slice of strings to the writer, each followed by a newline.
func WriteLines(ctx context.Context, writer io.Writer, lines []string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("write start: %w", err)
	}

	bufW := bufio.NewWriter(writer)

	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write cancelled: %w", err)
		}
		// skip unnecessary allocation
		if _, err := bufW.WriteString(line); err != nil {
			return fmt.Errorf("bufio write string: %w", err)
		}

		if err := bufW.WriteByte('\n'); err != nil {
			return fmt.Errorf("bufio write newline: %w", err)
		}
	}

	if err := bufW.Flush(); err != nil {
		return fmt.Errorf("bufio write flush: %w", err)
	}

	return nil
}
