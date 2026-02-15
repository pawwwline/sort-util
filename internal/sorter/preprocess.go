package sorter

import (
	"fmt"
	"sort-util/internal/config"
	"strconv"
)

// sortableRow used for preprocessing data before sorting, so it will be performed in O(N)
type sortableRow struct {
	original     string
	processedStr string  // result for getColumn and trimBlanks or whole string
	numKey       float64 // for -n (numeric)
	humanKey     int     // for -h (human suffix)
	monthKey     months  // for -M (months)

	isNumValid   bool
	isHumanValid bool
}

func newSortableRow(line string, cfg *config.Options) sortableRow {
	row := sortableRow{original: line}

	key := line
	if cfg.ColumnNum > 0 {
		key = getColumn(line, cfg.ColumnNum)
	}

	if cfg.TrailingBlanks {
		key = trimBlanks(key)
	}

	row.processedStr = key

	if cfg.Months {
		row.monthKey = parseMonths(key)
	}

	if cfg.HumanSuffix {
		val, err := parseHumanSuffix(key)
		row.humanKey = val
		row.isHumanValid = err == nil
	}

	if cfg.Numeric {
		val, err := strconv.ParseFloat(key, 64)
		row.numKey = val
		row.isNumValid = err == nil
	}

	return row
}

// getColumn returns the N-th column (1-based) delimited by tabs.
// and return empty string if column is not found
func getColumn(line string, columnNum int) string {
	// handle case of column with negative or 0 number
	if columnNum < 1 {
		return line
	}

	start := 0
	currentColumn := 1

	// finds start of column
	for currentColumn < columnNum {
		idx := -1

		for i := 0; i < len(line); i++ {
			if line[i] == '\t' {
				idx = i

				break
			}
		}
		if idx == -1 {
			return "" // didn't find out this column
		}

		start = idx + 1
		currentColumn++
	}

	// find out end of column
	end := start
	for end < len(line) && line[end] != '\t' {
		end++
	}

	return line[start:end]
}

// trimBlanks remove spaces by moving pointer without any allocation
func trimBlanks(line string) string {
	start := 0

	for start < len(line) && line[start] == ' ' {
		start++
	}

	end := len(line)
	for end > start && line[end-1] == ' ' {
		end--
	}

	return line[start:end]
}

// parseMonths parse months by bytes with 0 allocations.
func parseMonths(line string) months {
	if len(line) < minMonthLength {
		return 0
	}

	char1 := toUpper(line[0])
	char2 := toUpper(line[1])
	char3 := toUpper(line[2])

	switch char1 {
	case 'J':
		return parseJ(char2, char3) // JAN || JUN || JUL
	case 'M':
		return parseM(char2, char3) // MAR || MAY
	case 'A':
		return parseA(char2, char3) // APR || AUG
	default:
		return parseUnique(char1, char2, char3) // FEB || SEP || OCT || NOV || DEC
	}
}

func parseJ(char2, char3 byte) months {
	if char2 == 'A' && char3 == 'N' {
		return january
	}
	if char2 == 'U' {
		if char3 == 'N' {
			return june
		}
		if char3 == 'L' {
			return july
		}
	}

	return 0
}

func parseM(char2, char3 byte) months {
	if char2 == 'A' {
		if char3 == 'R' {
			return march
		}
		if char3 == 'Y' {
			return may
		}
	}

	return 0
}

func parseA(char2, char3 byte) months {
	if char2 == 'P' && char3 == 'R' {
		return april
	}
	if char2 == 'U' && char3 == 'G' {
		return august
	}

	return 0
}

// nolint:cyclop // parseUnique parses unique months by bytes to prevent allocations
func parseUnique(char1, char2, char3 byte) months {
	switch char1 {
	case 'F':
		if char2 == 'E' && char3 == 'B' {
			return february
		}
	case 'S':
		if char2 == 'E' && char3 == 'P' {
			return september
		}
	case 'O':
		if char2 == 'C' && char3 == 'T' {
			return october
		}
	case 'N':
		if char2 == 'O' && char3 == 'V' {
			return november
		}
	case 'D':
		if char2 == 'E' && char3 == 'C' {
			return december
		}
	}
	return 0
}

func toUpper(char byte) byte {
	if char >= 'a' && char <= 'z' {
		return char - ('a' - 'A')
	}
	return char
}

func parseHumanSuffix(line string) (int, error) {
	if len(line) == 0 {
		return 0, nil
	}

	suffix := line[len(line)-1]
	multiplier := 1

	switch suffix {
	case 'K':
		multiplier = kiB
		line = line[:len(line)-1]
	case 'M':
		multiplier = miB
		line = line[:len(line)-1]
	case 'G':
		multiplier = giB
		line = line[:len(line)-1]
	case 'T':
		multiplier = tiB
		line = line[:len(line)-1]
	}

	parsedNum, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("error parsing human suffix: %w", err)
	}

	return parsedNum * multiplier, nil
}
