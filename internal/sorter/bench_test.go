package sorter_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

func BenchmarkAutoSorter_Sort_FullMatrix(b *testing.B) {
	const size = 100000
	dataSets := map[string][]string{
		"Simple":      generateData(size, "banana", "apple", "cherry", "date"),
		"Numeric":     generateData(size, "100.5", "-10", "0", "25", "1000"),
		"HumanSuffix": generateData(size, "1K", "500M", "2.5G", "1T", "100"),
		"Months":      generateData(size, "Jan", "March", "February", "July"),
		"Columns":     generateData(size, "A\t10", "B\t5", "C\t100", "D\t1"),
		"Unique":      generateData(size, "duplicate", "original", "copy", "repeat"),
	}

	scenarios := []struct {
		name     string
		cfg      config.Options
		dataType string
	}{
		{"DefaultAlphabetical", config.Options{}, "Simple"},
		{"ReverseAlphabetical", config.Options{Reverse: true}, "Simple"},
		{"NumericSort", config.Options{Numeric: true}, "Numeric"},
		{"HumanReadable", config.Options{HumanSuffix: true}, "HumanSuffix"},
		{"MonthSort", config.Options{Months: true}, "Months"},
		{"ColumnNumeric", config.Options{ColumnNum: 2, Numeric: true}, "Columns"},
		{"UniqueOnly", config.Options{Unique: true}, "Simple"},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			srt := sorter.NewAutoSorter(&sc.cfg)
			rawContent := []byte(strings.Join(dataSets[sc.dataType], "\n"))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = srt.Sort(context.Background(), bytes.NewReader(rawContent), io.Discard)
			}
		})
	}
}

func BenchmarkAutoSorter_Sort_ExternalPath(b *testing.B) {
	const size = 1_000_000
	const thresholdBytes = 1 << 20 // 1 MiB → guarantees multiple runs

	data := generateData(size, "banana", "apple", "cherry", "date", "elderberry")
	rawContent := []byte(strings.Join(data, "\n"))

	scenarios := []struct {
		name string
		cfg  config.Options
	}{
		{"Default", config.Options{}},
		{"Numeric", config.Options{Numeric: true}},
		{"Reverse", config.Options{Reverse: true}},
		{"Unique", config.Options{Unique: true}},
	}

	for _, sc := range scenarios {
		b.Run(sc.name, func(b *testing.B) {
			srt := sorter.NewAutoSorter(&sc.cfg, sorter.WithThreshold(thresholdBytes))
			b.SetBytes(int64(len(rawContent)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = srt.Sort(context.Background(), bytes.NewReader(rawContent), io.Discard)
			}
		})
	}
}
