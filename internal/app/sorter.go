package app

import (
	"context"
	"io"
)

type Sorter interface {
	Sort(ctx context.Context, r io.Reader, w io.Writer) error
}
