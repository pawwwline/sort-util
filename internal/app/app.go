// Package app implements the high-level orchestration logic for the sorting utility.
package app

import (
	"context"
	"fmt"
	"io"
)

// App represents the core application responsible for orchestrating the sorting process.
type App struct {
	sorter Sorter
}

// New initializes and returns a new App instance with the provided Sorter implementation.
func New(sorter Sorter) *App {
	return &App{
		sorter: sorter,
	}
}

// Run orchestrates the sorting process by passing the reader and writer to the underlying sorter.
func (a *App) Run(ctx context.Context, r io.Reader, w io.Writer) error {
	if err := a.sorter.Sort(ctx, r, w); err != nil {
		return fmt.Errorf("sort input: %w", err)
	}

	return nil
}
