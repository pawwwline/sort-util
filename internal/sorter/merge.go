package sorter

import (
	"bufio"
	"container/heap"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"sort-util/internal/config"
)

// heapItem holds one line from a sorted run plus a scanner to fetch subsequent lines.
type heapItem struct {
	row     sortableRow
	scanner *bufio.Scanner
}

// runHeap implements heap.Interface for k-way merge.
type runHeap struct {
	items []*heapItem
	cfg   *config.Options
}

func (rh *runHeap) Len() int { return len(rh.items) }
func (rh *runHeap) Less(i, j int) bool {
	return compareForSort(&rh.items[i].row, &rh.items[j].row, rh.cfg) < 0
}
func (rh *runHeap) Swap(i, j int) { rh.items[i], rh.items[j] = rh.items[j], rh.items[i] }
func (rh *runHeap) Push(val any) {
	item, isValid := val.(*heapItem)
	if !isValid {
		panic("runHeap: unexpected push type")
	}
	rh.items = append(rh.items, item)
}
func (rh *runHeap) Pop() any {
	old := rh.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	rh.items = old[:n-1]
	return item
}

// merger performs k-way merge of pre-sorted run files into a single output stream.
type merger struct {
	cfg *config.Options
}

func newMerger(cfg *config.Options) *merger {
	return &merger{cfg: cfg}
}

// merge opens each file in paths, heap-merges their lines and writes the result to writer.
// Handles Unique deduplication and respects context cancellation.
func (m *merger) merge(ctx context.Context, paths []string, writer io.Writer) (retErr error) {
	files, runQueue, err := m.seedHeap(paths)
	if err != nil {
		return err
	}
	defer func() {
		for _, file := range files {
			if closeErr := file.Close(); closeErr != nil {
				retErr = errors.Join(retErr, fmt.Errorf("close run file: %w", closeErr))
			}
		}
	}()

	bufW := bufio.NewWriter(writer)

	var lastRow *sortableRow

	for runQueue.Len() > 0 {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("context cancelled: %w", ctxErr)
		}

		popped := heap.Pop(runQueue)
		item, isValid := popped.(*heapItem)
		if !isValid {
			return fmt.Errorf("unexpected heap item type")
		}

		var processErr error
		lastRow, processErr = m.processItem(bufW, item, lastRow)
		if processErr != nil {
			return processErr
		}

		if item.scanner.Scan() {
			item.row = newSortableRow(item.scanner.Text(), m.cfg)
			heap.Push(runQueue, item)
		}
	}

	if err := bufW.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	return nil
}

// seedHeap opens each run file, initialises a bufio.Scanner for it, seeds the heap
// with the first line of each file, and returns the opened files and the ready heap.
func (m *merger) seedHeap(paths []string) ([]*os.File, *runHeap, error) {
	files := make([]*os.File, 0, len(paths))
	runQueue := &runHeap{cfg: m.cfg, items: make([]*heapItem, 0, len(paths))}

	for _, path := range paths {
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			openErr := fmt.Errorf("open temp file: %w", err)
			var closeErrs []error
			for _, openedFile := range files {
				if closeErr := openedFile.Close(); closeErr != nil {
					closeErrs = append(closeErrs, closeErr)
				}
			}
			return nil, nil, errors.Join(append([]error{openErr}, closeErrs...)...)
		}
		files = append(files, file)

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, scannerInitBufSize)
		scanner.Buffer(buf, defaultMaxLineSize)

		if scanner.Scan() {
			runQueue.items = append(runQueue.items, &heapItem{
				row:     newSortableRow(scanner.Text(), m.cfg),
				scanner: scanner,
			})
		}
	}

	heap.Init(runQueue)
	return files, runQueue, nil
}

// processItem writes item to bufW unless it is a duplicate of lastRow.
// Returns the updated lastRow pointer and any write error.
func (m *merger) processItem(bufW *bufio.Writer, item *heapItem, lastRow *sortableRow) (*sortableRow, error) {
	if m.isDuplicate(item, lastRow) {
		return lastRow, nil
	}
	if err := writeRow(bufW, item.row.original); err != nil {
		return nil, err
	}
	rowCopy := item.row
	return &rowCopy, nil
}

func (m *merger) isDuplicate(item *heapItem, lastRow *sortableRow) bool {
	return m.cfg.Unique && lastRow != nil && compare(&item.row, lastRow, m.cfg) == 0
}

func writeRow(bufW *bufio.Writer, line string) error {
	if _, err := bufW.WriteString(line); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if err := bufW.WriteByte('\n'); err != nil {
		return fmt.Errorf("write output newline: %w", err)
	}
	return nil
}
