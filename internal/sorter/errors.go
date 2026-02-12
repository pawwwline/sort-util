package sorter

import "errors"

var (
	// ErrNotSorted checks if provided strings was sorted
	ErrNotSorted = errors.New("not sorted")
)
