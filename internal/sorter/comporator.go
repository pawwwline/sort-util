package sorter

import (
	"strconv"

	"sort-util/internal/config"
)

func newComparator(cfg config.Options) func(a, b string) bool {
	return func(a, b string) bool {
		var isLess bool

		if cfg.Numeric {
			isLess = compareNumeric(a, b)
		} else {
			return a < b
		}

		if cfg.Reverse {
			return !isLess
		}

		return isLess
	}
}

func compareNumeric(a, b string) bool {
	//ignore error so strings that not numeric will be interpreted as 0.0
	v1, _ := strconv.ParseFloat(a, 64)
	v2, _ := strconv.ParseFloat(b, 64)

	if v1 != v2 {
		return v1 < v2
	}

	return a < b
}

// uniqueLines in-place deleting non unique elements
func uniqueLines(lines []string) []string {
	if len(lines) < 2 {
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
