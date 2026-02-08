package app

import (
	"context"
	"io"
)

// Sorter defines the contract for implementing various sorting algorithms.
type Sorter interface {
	// Sort reads data from r, performs sorting, and writes the result to w.
	Sort(ctx context.Context, r io.Reader, w io.Writer) error
}
