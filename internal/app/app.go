package app

import (
	"context"
	"io"
)

type App struct {
	sorter Sorter
}

func New(sorter Sorter) *App {
	return &App{
		sorter: sorter,
	}
}

func (a *App) Run(ctx context.Context, r io.Reader, w io.Writer) error {
	return a.sorter.Sort(ctx, r, w)
}
