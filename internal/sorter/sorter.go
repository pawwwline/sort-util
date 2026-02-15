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
	cfg *config.Options
}

// NewInMemory initializes a new InMemory sorter with the provided configuration options.
func NewInMemory(cfg *config.Options) *InMemory {
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

func (i *InMemory) sortLines(lines []string) {
	rows := make([]sortableRow, len(lines))

	// preprocess lines O(N)
	for idx, line := range lines {
		rows[idx] = newSortableRow(line, i.cfg)
	}

	slices.SortFunc(rows, func(a, b sortableRow) int {
		return compare(&a, &b, i.cfg)
	})

	for idx, row := range rows {
		lines[idx] = row.original
	}

	if i.cfg.Reverse {
		slices.Reverse(lines)
	}
}

// uniqueLines in-place deleting non unique elements using two pointers approach
func uniqueLines(lines []string) []string {
	minLines := 2

	if len(lines) < minLines {
		return lines
	}

	// slow tracks the last unique element found
	slow := 0
	for fast := 1; fast < len(lines); fast++ {
		if lines[fast] != lines[slow] {
			slow++
			lines[slow] = lines[fast]
		}
	}

	return lines[:slow+1]
}
