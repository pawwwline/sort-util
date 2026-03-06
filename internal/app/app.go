// Package app implements the high-level orchestration logic for the sorting utility.
package app

import (
	"context"
	"fmt"
	"io"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

// App represents the core application responsible for orchestrating the sorting process.
type App struct {
	cfg *config.Options
}

// New initializes and returns a new App instance with the provided Sorter implementation.
func New(cfg *config.Options) *App {
	return &App{
		cfg: cfg,
	}
}

// Run orchestrates the sorting process by passing the reader and writer to the underlying sorter.
func (a *App) Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	if a.cfg.CheckSorted {
		checker := sorter.NewChecker(a.cfg)
		err := checker.CheckSorted(ctx, reader)
		if err != nil {
			return fmt.Errorf("check sorting: %w", err)
		}

		return nil
	}

	autoSorter := sorter.NewAutoSorter(a.cfg)
	if err := autoSorter.Sort(ctx, reader, writer); err != nil {
		return fmt.Errorf("sort input: %w", err)
	}

	return nil
}
