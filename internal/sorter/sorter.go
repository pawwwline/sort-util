package sorter

import (
	"context"
	"io"
	"slices"

	"sort-util/internal/config"
	"sort-util/internal/provider"
)

type InMemory struct {
	cfg config.Options
}

func NewInMemory(cfg config.Options) *InMemory {
	return &InMemory{cfg: cfg}
}

func (i *InMemory) Sort(ctx context.Context, r io.Reader, w io.Writer) error {
	strings, err := provider.ReadLines(ctx, r)
	if err != nil {
		return err
	}

	i.sortLines(strings) //sort lines in place

	if i.cfg.Unique {
		strings = uniqueLines(strings)
	}

	err = provider.WriteLines(ctx, w, strings)
	if err != nil {
		return err
	}

	return nil
}

func (i *InMemory) sortLines(sortedStrings []string) {
	compare := newComparator(i.cfg)

	slices.SortFunc(sortedStrings, func(a, b string) int {
		if compare(a, b) {
			return -1 // a is less
		}
		if compare(b, a) {
			return 1 // b is less
		}
		return 0 // equal
	})
}
