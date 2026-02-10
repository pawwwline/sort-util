// Package config defines the configuration structures for the application settings.
package config

// Options defines the configuration flags for sorting behavior.
type Options struct {
	Reverse        bool
	Numeric        bool
	Unique         bool
	Sorted         bool
	TrailingBlanks bool
	CheckSorted    bool
	Months         bool
	ColumnNum      int
}
