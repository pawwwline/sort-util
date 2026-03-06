package sorter

import (
	"context"
	"fmt"
	"io"

	"sort-util/internal/config"
	"sort-util/internal/provider"
)

// AutoSorter transparently switches between in-memory and external sorting depending
// on how much data has been read. If the input fits within memoryThreshold it never
// touches the disk.
type AutoSorter struct {
	cfg *config.Options
}

// NewAutoSorter initializes a new AutoSorter with the provided configuration.
func NewAutoSorter(cfg *config.Options) *AutoSorter {
	return &AutoSorter{cfg: cfg}
}

// Sort reads the input and delegates to in-memory or external sort as needed.
func (a *AutoSorter) Sort(ctx context.Context, reader io.Reader, writer io.Writer) error {
	lineReader := provider.NewLineReader(reader, defaultMaxLineSize)

	var buffer []string
	var memUsed int

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		line, err := lineReader.Next()
		if err == io.EOF {
			return a.runInMemory(ctx, buffer, writer)
		}
		if err != nil {
			return fmt.Errorf("read line: %w", err)
		}

		buffer = append(buffer, line)
		memUsed += len(line)

		if memUsed > memoryThreshold {
			return a.runExternal(ctx, buffer, lineReader, writer)
		}
	}
}

func (a *AutoSorter) runInMemory(ctx context.Context, buffer []string, writer io.Writer) error {
	mem := NewInMemory(a.cfg)
	buffer = mem.sortLines(buffer)

	if err := provider.WriteLines(ctx, writer, buffer); err != nil {
		return fmt.Errorf("write lines: %w", err)
	}

	return nil
}

func (a *AutoSorter) runExternal(
	ctx context.Context,
	buffer []string,
	lineReader *provider.ScannerLineReader,
	writer io.Writer,
) error {
	ext := NewExternal(a.cfg)
	return ext.sortCore(ctx, buffer, lineReader, writer)
}
