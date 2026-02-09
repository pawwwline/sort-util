package sorter

import (
	"strconv"

	"sort-util/internal/config"
)

func newComparator(cfg config.Options) func(str1, str2 string) bool {
	return func(str1, str2 string) bool {
		if cfg.TrailingBlanks {
			str1 = trimBlanks(str1)
			str2 = trimBlanks(str2)
		}

		var isLess bool

		if cfg.Numeric {
			isLess = compareNumeric(str1, str2)
		} else {
			isLess = str1 < str2
		}

		if cfg.Months {
			isLess = compareMonths(str1, str2)
		}

		if cfg.Reverse {
			isLess = !isLess
		}

		return isLess
	}
}

func compareNumeric(str1, str2 string) bool {
	// ignore error so strings that not numeric will be interpreted as 0.0
	v1, err1 := strconv.ParseFloat(str1, 64)
	v2, err2 := strconv.ParseFloat(str2, 64)

	if err1 != nil && err2 != nil {
		return str1 < str2
	}

	// if num is not numeric it is always more
	if err1 != nil {
		return false // is more
	}
	if err2 != nil {
		return true // is less
	}

	if v1 != v2 {
		return v1 < v2
	}

	return str1 < str2
}

// uniqueLines in-place deleting non unique elements using two pointers approach
func uniqueLines(lines []string) []string {
	minLines := 2

	if len(lines) < minLines {
		return lines
	}

	// slow tracks the last unique element found
	slow := 0
	for fast := 1; fast < len(lines); fast++ {
		if lines[fast] != lines[slow] {
			slow++
			lines[slow] = lines[fast]
		}
	}

	return lines[:slow+1]
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

func compareMonths(str1, str2 string) bool {
	month1 := parseMonths(str1)
	month2 := parseMonths(str2)

	return month1 < month2
}

// parseMonths parse months by bytes with 0 allocations.
func parseMonths(line string) months {
	if len(line) < 3 {
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

// nolint p
func parseUnique(c1, c2, c3 byte) months {
	switch c1 {
	case 'F':
		if c2 == 'E' && c3 == 'B' {
			return february
		}
	case 'S':
		if c2 == 'E' && c3 == 'P' {
			return september
		}
	case 'O':
		if c2 == 'C' && c3 == 'T' {
			return october
		}
	case 'N':
		if c2 == 'O' && c3 == 'V' {
			return november
		}
	case 'D':
		if c2 == 'E' && c3 == 'C' {
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
