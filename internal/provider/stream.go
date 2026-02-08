// Package provider implements I/O utilities for reading and writing data streams.
package provider

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

// ReadLines reads all lines from the provided reader into a string slice.
func ReadLines(ctx context.Context, reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	// add context checking to exit earlier
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled %w", ctx.Err())
		default:
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return lines, nil
}

// WriteLines writes a slice of strings to the writer, each followed by a newline.
func WriteLines(ctx context.Context, writer io.Writer, lines []string) error {
	bufW := bufio.NewWriter(writer)

	defer func() { _ = bufW.Flush() }()

	for _, line := range lines {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled %w", ctx.Err())
		default:
			if _, err := bufW.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("bufio write: %w", err)
			}
		}
	}

	return nil
}
