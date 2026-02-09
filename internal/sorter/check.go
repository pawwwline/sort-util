package sorter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"sort-util/internal/config"
)

type Checker struct {
	cfg config.Options
}

func NewChecker(cfg config.Options) *Checker {
	return &Checker{cfg: cfg}
}

// CheckSorted use bufio scanner and check only prev and next lines
func (c *Checker) CheckSorted(ctx context.Context, reader io.Reader) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("write start: %w", err)
	}

	scanner := bufio.NewScanner(reader)
	compare := newComparator(c.cfg)

	var prev string
	if scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write cancelled: %w", err)
		}
		prev = scanner.Text()
	}

	lineNum := 1
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("write cancelled: %w", err)
		}
		lineNum++
		curr := scanner.Text()

		if compare(curr, prev) {
			_, err := fmt.Fprintf(os.Stderr, "sort: -:%d: disorder: %s\n", lineNum, curr)
			if err != nil {
				return fmt.Errorf("error writing description: %w", err)
			}

			return ErrNotSorted
		}
		prev = curr
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read lines: %w", err)
	}

	return nil
}
