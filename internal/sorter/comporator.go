package sorter

import (
	"strconv"

	"sort-util/internal/config"
)

func newComparator(cfg config.Options) func(str1, str2 string) bool {
	return func(str1, str2 string) bool {
		var isLess bool

		if cfg.Numeric {
			isLess = compareNumeric(str1, str2)
		} else {
			return str1 < str2
		}

		if cfg.Reverse {
			return !isLess
		}

		return isLess
	}
}

func compareNumeric(str1, str2 string) bool {
	// ignore error so strings that not numeric will be interpreted as 0.0
	v1, _ := strconv.ParseFloat(str1, 64)
	v2, _ := strconv.ParseFloat(str2, 64)

	if v1 != v2 {
		return v1 < v2
	}

	return str1 < str2
}

// uniqueLines in-place deleting non unique elements
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
