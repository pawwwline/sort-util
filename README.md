# sort-util

GNU `sort`-compatible text sorting utility written in Go.

Reads lines from a file or stdin, sorts them according to the specified flags, and writes the result to stdout. Automatically switches to disk-backed external merge sort when the input exceeds 256 MiB, so arbitrarily large files are handled without running out of memory.

---

## Installation

Build from source:

```bash
git clone <repo>
cd sort-util
go build -o sort-util ./cmd/sort-util
```

---

## Usage

```
sort-util [flags] [file]
```

If `file` is omitted, input is read from stdin.

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--numeric` | `-n` | false | Sort by numeric value (supports integers and floats) |
| `--reverse` | `-r` | false | Reverse the sort order |
| `--unique` | `-u` | false | Output only the first of equal lines |
| `--blanks` | `-b` | false | Ignore leading and trailing whitespace when comparing |
| `--months` | `-m` | false | Sort by month name (Jan < Feb < … < Dec) |
| `--human-numeric-sort` | | false | Sort by numeric value with SI suffixes (K, M, G, T) |
| `--column` | `-k` | 1 | Sort by the Nth tab-delimited column |
| `--check-sorted` | `-c` | false | Check whether input is sorted; exit non-zero if not |

---

## Examples

**Alphabetical sort (default)**
```bash
sort-util input.txt
```

**Numeric sort**
```bash
echo -e "10\n2\n1" | sort-util -n
# 1
# 2
# 10
```

**Reverse sort**
```bash
echo -e "apple\ncherry\nbanana" | sort-util -r
# cherry
# banana
# apple
```

**Remove duplicates**
```bash
echo -e "apple\nbanana\napple" | sort-util -u
# apple
# banana
```

**Sort by month**
```bash
echo -e "March\nJanuary\nFebruary" | sort-util -m
# January
# February
# March
```

**Human-readable sizes**
```bash
echo -e "2G\n500M\n1T\n100K" | sort-util --human-numeric-sort
# 100K
# 500M
# 2G
# 1T
```

**Sort by second column (tab-separated)**
```bash
echo -e "B\t30\nA\t5\nC\t100" | sort-util -k 2 -n
# A	5
# B	30
# C	100
```

**Check if file is already sorted**
```bash
sort-util -c input.txt && echo "sorted" || echo "not sorted"
```

---

## How it works

`sort-util` uses an `AutoSorter` that reads input line by line and tracks memory consumption:

- **≤ 256 MiB** — all lines are sorted in memory using `slices.SortFunc`.
- **> 256 MiB** — switches transparently to **external merge sort**: lines are sorted in chunks and flushed to temporary files, then merged using a k-way min-heap. Temporary files are removed automatically after the merge completes.

Deduplication (`-u`) is applied during the final merge pass, so it works correctly on both paths without buffering the entire output.

---

## Development

```bash
# Run all tests
go test ./...

# Run unit tests with verbose output
go test ./internal/sorter/... -v -run TestInMemory
go test ./internal/sorter/... -v -run TestAutoSorter
go test ./internal/sorter/... -v -run TestExternal

# Run benchmarks (in-memory path, 100K lines)
go test ./internal/sorter/... -bench=BenchmarkInMemory -benchmem -benchtime=5s

# Run benchmarks (AutoSorter — both in-memory and external paths)
go test ./internal/sorter/... -bench=BenchmarkAutoSorter -benchmem -benchtime=5s
```

### Project structure

```
cmd/
  sort-util/main.go      — entry point
  root.go                — CLI flags (cobra)
internal/
  config/options.go      — Options struct
  app/app.go             — orchestration (CheckSorted or AutoSorter)
  provider/stream.go     — streaming line reader and writer
  sorter/
    sorter.go            — InMemory sorter
    autosorter.go        — AutoSorter + Option / WithThreshold
    external.go          — External (disk-backed) sorter
    merge.go             — k-way heap merger
    comporator.go        — compare / compareForSort
    preprocess.go        — sortableRow preprocessing
    check.go             — Checker (sorted-order validation)
```

---

## Requirements

- Go 1.25.5
