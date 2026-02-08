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
			isLess = str1 < str2
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
