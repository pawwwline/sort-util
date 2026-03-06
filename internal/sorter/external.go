package sorter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"sort-util/internal/config"
	"sort-util/internal/provider"
)

// External implements disk-backed external merge sort for datasets that exceed available memory.
type External struct {
	cfg       *config.Options
	threshold int
}

// NewExternal initializes a new External sorter with the provided configuration.
func NewExternal(cfg *config.Options, opts ...Option) *External {
	o := &sorterOptions{threshold: memoryThreshold}
	for _, opt := range opts {
		opt(o)
	}
	return &External{cfg: cfg, threshold: o.threshold}
}

// Sort reads all input, flushing sorted runs to disk when the memory threshold is
// exceeded, then k-way merges all runs into the output.
func (e *External) Sort(ctx context.Context, reader io.Reader, writer io.Writer) error {
	lineReader := provider.NewLineReader(reader, defaultMaxLineSize)
	return e.sortCore(ctx, nil, lineReader, writer)
}

// sortCore is the shared implementation.
// AutoSorter calls this directly, passing the already-read initialBatch plus the
// remaining line reader so we avoid re-reading data that is already in memory.
func (e *External) sortCore(
	ctx context.Context,
	initialBatch []string,
	lineReader *provider.ScannerLineReader,
	writer io.Writer,
) (retErr error) {
	var tempPaths []string
	defer func() {
		for _, path := range tempPaths {
			if removeErr := os.Remove(path); removeErr != nil && retErr == nil {
				retErr = fmt.Errorf("remove temp file: %w", removeErr)
			}
		}
	}()

	var remaining []string
	tempPaths, remaining, retErr = e.createRuns(ctx, initialBatch, lineReader)
	if retErr != nil {
		return
	}

	if len(tempPaths) == 0 {
		return e.sortInMemory(ctx, remaining, writer)
	}

	if len(remaining) > 0 {
		path, flushErr := e.sortAndFlush(remaining)
		if flushErr != nil {
			return flushErr
		}
		tempPaths = append(tempPaths, path)
	}

	return newMerger(e.cfg).merge(ctx, tempPaths, writer)
}

// createRuns reads lines from lineReader (prepending any initialBatch), flushing sorted
// runs to disk whenever the memory threshold is reached.
// It returns the list of created temp file paths and the leftover batch not yet flushed.
// On error it still returns the paths created so far so the caller can clean them up.
func (e *External) createRuns(
	ctx context.Context,
	initialBatch []string,
	lineReader *provider.ScannerLineReader,
) (paths []string, remaining []string, err error) {
	batch := initialBatch
	batchSize := batchByteSize(initialBatch)

	if len(batch) > 0 && batchSize >= e.threshold {
		path, flushErr := e.sortAndFlush(batch)
		if flushErr != nil {
			return paths, nil, flushErr
		}
		paths = append(paths, path)
		batch, batchSize = nil, 0
	}

	for {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return paths, nil, fmt.Errorf("context cancelled: %w", ctxErr)
		}

		line, lineErr := lineReader.Next()
		if lineErr == io.EOF {
			return paths, batch, nil
		}
		if lineErr != nil {
			return paths, nil, fmt.Errorf("read line: %w", lineErr)
		}

		batch = append(batch, line)
		batchSize += len(line)

		if batchSize >= e.threshold {
			path, flushErr := e.sortAndFlush(batch)
			if flushErr != nil {
				return paths, nil, flushErr
			}
			paths = append(paths, path)
			batch, batchSize = nil, 0
		}
	}
}

func batchByteSize(batch []string) int {
	total := 0
	for _, line := range batch {
		total += len(line)
	}
	return total
}

func (e *External) sortInMemory(ctx context.Context, lines []string, writer io.Writer) error {
	mem := &InMemory{cfg: e.cfg}
	lines = mem.sortLines(lines)

	if err := provider.WriteLines(ctx, writer, lines); err != nil {
		return fmt.Errorf("write lines: %w", err)
	}

	return nil
}

func (e *External) sortedRows(lines []string) []sortableRow {
	rows := make([]sortableRow, len(lines))
	for idx, line := range lines {
		rows[idx] = newSortableRow(line, e.cfg)
	}
	slices.SortFunc(rows, func(rowA, rowB sortableRow) int {
		return compareForSort(&rowA, &rowB, e.cfg)
	})
	return rows
}

func (e *External) sortAndFlush(lines []string) (string, error) {
	rows := e.sortedRows(lines)

	tmpFile, err := os.CreateTemp("", "sort-util-run-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	name := tmpFile.Name()
	bufW := bufio.NewWriter(tmpFile)

	for _, row := range rows {
		if _, werr := bufW.WriteString(row.original); werr != nil {
			return "", fmt.Errorf("write temp: %w", errors.Join(werr, tmpFile.Close()))
		}
		if werr := bufW.WriteByte('\n'); werr != nil {
			return "", fmt.Errorf("write temp newline: %w", errors.Join(werr, tmpFile.Close()))
		}
	}

	if werr := bufW.Flush(); werr != nil {
		return "", fmt.Errorf("flush temp: %w", errors.Join(werr, tmpFile.Close()))
	}

	if cerr := tmpFile.Close(); cerr != nil {
		return "", fmt.Errorf("close temp: %w", cerr)
	}

	return name, nil
}
