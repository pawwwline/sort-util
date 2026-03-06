// Package sorter provides implementations of various sorting strategies.
package sorter

import (
	"context"
	"fmt"
	"io"
	"slices"
	"sort-util/internal/provider"

	"sort-util/internal/config"
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
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	lineReader := provider.NewLineReader(reader, defaultMaxLineSize)

	var lines []string

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		line, err := lineReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read line: %w", err)
		}

		lines = append(lines, line)
	}

	lines = i.sortLines(lines)

	if err := provider.WriteLines(ctx, writer, lines); err != nil {
		return fmt.Errorf("write lines: %w", err)
	}

	return nil
}

// sortLines preprocesses, sorts and optionally deduplicates lines in-place.
// It returns a (possibly shorter) slice — callers must use the returned value.
func (i *InMemory) sortLines(lines []string) []string {
	rows := make([]sortableRow, len(lines))

	// preprocess lines O(N)
	for idx, line := range lines {
		rows[idx] = newSortableRow(line, i.cfg)
	}

	slices.SortFunc(rows, func(a, b sortableRow) int {
		return compareForSort(&a, &b, i.cfg)
	})

	if i.cfg.Unique {
		rows = uniqueRows(rows, i.cfg)
	}

	lines = lines[:len(rows)]
	for idx, row := range rows {
		lines[idx] = row.original
	}

	return lines
}

// uniqueRows deduplicates sorted rows by sort key using a two-pointer approach.
func uniqueRows(rows []sortableRow, cfg *config.Options) []sortableRow {
	if len(rows) < minDeduplicateLen {
		return rows
	}

	slow := 0
	for fast := 1; fast < len(rows); fast++ {
		if compare(&rows[slow], &rows[fast], cfg) != 0 {
			slow++
			rows[slow] = rows[fast]
		}
	}

	return rows[:slow+1]
}
