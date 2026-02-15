package sorter_test

import (
	"bytes"
	"context"
	"io"
	"math/rand/v2"
	"strings"
	"testing"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

// nolint:funlen,varnamelen
func TestInMemory_Sort(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name     string
		cfg      config.Options
		ctx      context.Context
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Empty input",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Single line input",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "string\n",
			expected: "string\n",
			wantErr:  false,
		},
		{
			name:     "Default alphabetical sort",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "banana\napple\ncherry\n",
			expected: "apple\nbanana\ncherry\n",
			wantErr:  false,
		},
		{
			name:     "Sort with empty lines in between",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "banana\n\napple\n",
			expected: "\napple\nbanana\n",
			wantErr:  false,
		},
		{
			name:     "Reverse alphabetical sort",
			cfg:      config.Options{Reverse: true},
			ctx:      context.Background(),
			input:    "apple\nbanana\ncherry\n",
			expected: "cherry\nbanana\napple\n",
			wantErr:  false,
		},
		{
			name:     "Numeric sort",
			cfg:      config.Options{Numeric: true},
			input:    "10\n2\n1\n",
			ctx:      context.Background(),
			expected: "1\n2\n10\n",
			wantErr:  false,
		},
		{
			name:     "Stability test (equal numeric values)",
			cfg:      config.Options{Numeric: true},
			ctx:      context.Background(),
			input:    "02\n2\n002\n",
			expected: "002\n02\n2\n",
			wantErr:  false,
		},
		{
			name:     "Numeric sort and reverse",
			cfg:      config.Options{Numeric: true, Reverse: true},
			ctx:      context.Background(),
			input:    "10\n20\n1\n",
			expected: "20\n10\n1\n",
			wantErr:  false,
		},
		{
			name:     "Numeric with negative and floating points",
			cfg:      config.Options{Numeric: true},
			ctx:      context.Background(),
			input:    "10.5\n-1\n0\n2\n",
			expected: "-1\n0\n2\n10.5\n",
			wantErr:  false,
		},
		{
			name: "Mixed blanks and numeric",
			cfg: config.Options{
				TrailingBlanks: true,
				Numeric:        true,
			},
			ctx:      context.Background(),
			input:    "  10\n 2\n",
			expected: " 2\n  10\n",
		},
		{
			name:     "Numeric sort has alphabetical chars",
			cfg:      config.Options{Numeric: true},
			ctx:      context.Background(),
			input:    "apple\nbanana\n1\n0.5\n",
			expected: "0.5\n1\napple\nbanana\n",
			wantErr:  false,
		},
		{
			name:     "Numeric sort has alphabetical chars alphabetical are sorted",
			cfg:      config.Options{Numeric: true},
			ctx:      context.Background(),
			input:    "banana\napple\n0.5\n",
			expected: "0.5\napple\nbanana\n",
			wantErr:  false,
		},
		{
			name:     "Unique lines",
			cfg:      config.Options{Unique: true},
			ctx:      context.Background(),
			input:    "apple\nbanana\napple\n",
			expected: "apple\nbanana\n",
			wantErr:  false,
		},
		{
			name: "Ignore leading blanks (spaces)",
			cfg: config.Options{
				TrailingBlanks: true,
			},
			ctx:      context.Background(),
			input:    "  b\na\n",
			expected: "a\n  b\n",
		},
		{
			name:    "Test context canceled",
			cfg:     config.Options{},
			ctx:     cancelledCtx,
			input:   "b\na\n",
			wantErr: true,
		},
		{
			name: "Basic chronological months sort",
			cfg: config.Options{
				Months: true,
			},
			ctx:      context.Background(),
			input:    "march 3rd line\nJAN 1st line\nFeb 2nd line\n",
			expected: "JAN 1st line\nFeb 2nd line\nmarch 3rd line\n",
			wantErr:  false,
		}, {name: "Full months names and mixed case",
			cfg: config.Options{
				Months: true,
			},
			ctx:      context.Background(),
			input:    "July data\njanuary data\nMAY data\n",
			expected: "january data\nMAY data\nJuly data\n",
			wantErr:  false},
		{name: "Full months names and mixed case and reverse",
			cfg: config.Options{
				Months:  true,
				Reverse: true,
			},
			ctx:      context.Background(),
			input:    "July data\njanuary data\nMAY data\n",
			expected: "July data\nMAY data\njanuary data\n",
			wantErr:  false},
		{
			name:     "Column sorting with numeric values",
			cfg:      config.Options{ColumnNum: 2, Numeric: true},
			ctx:      context.Background(),
			input:    "ID_B\t100\nID_A\t20\nID_C\t5\n",
			expected: "ID_C\t5\nID_A\t20\nID_B\t100\n",
		},
		{
			name:     "Missing column (fallback to empty string)",
			cfg:      config.Options{ColumnNum: 5, Numeric: true},
			ctx:      context.Background(),
			input:    "A\t10\nB\t20\n",
			expected: "A\t10\nB\t20\n",
		},
		{
			name:     "Leading spaces in column data",
			cfg:      config.Options{ColumnNum: 2, Numeric: true, TrailingBlanks: true},
			input:    "item1\t 50\nitem2\t 10\n",
			expected: "item2\t 10\nitem1\t 50\n",
			ctx:      context.Background(),
		},
		{
			name:     "Human suffix sort",
			cfg:      config.Options{HumanSuffix: true},
			ctx:      context.Background(),
			input:    "2000T\n10\n1000000M\n",
			expected: "10\n1000000M\n2000T\n",
		},
		{
			name:     "Human suffix sort with not numeric values",
			cfg:      config.Options{HumanSuffix: true},
			ctx:      context.Background(),
			input:    "2000T\n10\n1000000M\nbanana\napple\n",
			expected: "10\n1000000M\n2000T\napple\nbanana\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sorter.NewInMemory(&tt.cfg)
			reader := strings.NewReader(tt.input)
			writer := &bytes.Buffer{}

			err := s.Sort(tt.ctx, reader, writer)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Sort() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "context canceled") {
					t.Errorf("expected cancellation error, got: %v", err)
				}
				return
			}

			if writer.String() != tt.expected {
				t.Errorf("got:\n%q\nwant:\n%q", writer.String(), tt.expected)
			}
		})
	}
}

// nolint:gosec // generateData generate random data with same seed for benchmark tests
func generateData(n int, samples ...string) []string {
	r := rand.New(rand.NewPCG(42, 1024))
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = samples[r.IntN(len(samples))]
	}
	return res
}

func BenchmarkInMemory_Sort_FullMatrix(b *testing.B) {
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
			sorterFunc := sorter.NewInMemory(&sc.cfg)

			rawContent := []byte(strings.Join(dataSets[sc.dataType], "\n"))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = sorterFunc.Sort(context.Background(), bytes.NewReader(rawContent), io.Discard)
			}
		})
	}
}
