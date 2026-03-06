package sorter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"sort-util/internal/config"
)

// Checker provides methods to determine the sort order of data strings.
type Checker struct {
	cfg *config.Options
}

// NewChecker provides new Checker struct
func NewChecker(cfg *config.Options) *Checker {
	return &Checker{cfg: cfg}
}

// CheckSorted use bufio scanner and check only prev and next lines
func (c *Checker) CheckSorted(ctx context.Context, reader io.Reader) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	isFirst := true

	var prevRow sortableRow

	for scanner.Scan() {
		lineNum++
		currText := scanner.Text()
		currRow := newSortableRow(currText, c.cfg)

		if isFirst {
			prevRow = currRow
			isFirst = false
			continue
		}

		res := compareForSort(&currRow, &prevRow, c.cfg)

		isLess := res < 0
		isDuplicate := c.cfg.Unique && res == 0

		if isLess || isDuplicate {
			return c.reportDisorder(lineNum, currText)
		}

		prevRow = currRow
		isFirst = false
	}

	err := scanner.Err()
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

func (c *Checker) reportDisorder(line int, text string) error {
	_, err := fmt.Fprintf(os.Stderr, "sort: -:%d: disorder: %s\n", line, text)
	if err != nil {
		return fmt.Errorf("error writing to stderr: %w", err)
	}
	return ErrNotSorted
}
