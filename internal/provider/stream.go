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
	// TODO: remove this check to tool function
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read start: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(reader)

	// add context checking to exit earlier
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("reading: %w", err)
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return lines, nil
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
