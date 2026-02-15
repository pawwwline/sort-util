package sorter

import (
	"cmp"

	"sort-util/internal/config"
)

func compare(rowA, rowB *sortableRow, cfg *config.Options) int {
	if cfg.Months {
		if c := cmp.Compare(rowA.monthKey, rowB.monthKey); c != 0 {
			return c
		}
	}

	if cfg.HumanSuffix {
		if res, handled := cmpValid(rowA.humanKey, rowA.isHumanValid, rowB.humanKey, rowB.isHumanValid); handled {
			return res
		}
	}

	if cfg.Numeric {
		if res, handled := cmpValid(rowA.numKey, rowA.isNumValid, rowB.numKey, rowB.isNumValid); handled {
			return res
		}
	}

	return cmp.Compare(rowA.processedStr, rowB.processedStr)
}

// cmpValid order
func cmpValid[T cmp.Ordered](aVal T, aOk bool, bVal T, bOk bool) (int, bool) {
	if !aOk && !bOk {
		return 0, false
	} // both invalid
	if !aOk {
		return 1, true
	} // a invalid >> a is more
	if !bOk {
		return -1, true
	} // b invalid >> a is less

	// if both valid compare
	res := cmp.Compare(aVal, bVal)
	if res == 0 {
		return 0, false // equal
	}

	return res, true
}
