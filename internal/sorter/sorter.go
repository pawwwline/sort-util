// Package sorter provides implementations of various sorting strategies.
package sorter

import (
	"context"
	"fmt"
	"io"
	"slices"

	"sort-util/internal/config"
	"sort-util/internal/provider"
)

// InMemory provides an implementation of the Sorter interface that processes data in RAM.
type InMemory struct {
	cfg config.Options
}

// NewInMemory initializes a new InMemory sorter with the provided configuration options.
func NewInMemory(cfg config.Options) *InMemory {
	return &InMemory{cfg: cfg}
}

// Sort in memory implementation with config flag handling.
func (i *InMemory) Sort(ctx context.Context, reader io.Reader, writer io.Writer) error {
	lines, err := provider.ReadLines(ctx, reader)
	if err != nil {
		return fmt.Errorf("read lines: %w", err)
	}

	i.sortLines(lines) // sort lines in place

	if i.cfg.Unique {
		lines = uniqueLines(lines)
	}

	err = provider.WriteLines(ctx, writer, lines)
	if err != nil {
		return fmt.Errorf("write lines: %w", err)
	}

	return nil
}

func (i *InMemory) sortLines(sortedLines []string) {
	compare := newComparator(i.cfg)

	slices.SortFunc(sortedLines, func(a, b string) int {
		if compare(a, b) {
			return -1 // a is less
		}
		if compare(b, a) {
			return 1 // b is less
		}
		return 0 // equal
	})
}
