package sorter_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

// smallThreshold forces the external sort path on tiny inputs (bytes).
const smallThreshold = 50

// nolint:funlen
func TestAutoSorter_Sort(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name      string
		cfg       config.Options
		ctx       context.Context
		input     string
		expected  string
		threshold int // 0 means default (in-memory path for small inputs)
		wantErr   bool
	}{
		// --- in-memory path (default threshold, small input) ---
		{
			name:     "Empty input",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "",
			expected: "",
		},
		{
			name:     "Single line",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "hello\n",
			expected: "hello\n",
		},
		{
			name:     "Alphabetical in-memory",
			cfg:      config.Options{},
			ctx:      context.Background(),
			input:    "banana\napple\ncherry\n",
			expected: "apple\nbanana\ncherry\n",
		},
		{
			name:     "Reverse in-memory",
			cfg:      config.Options{Reverse: true},
			ctx:      context.Background(),
			input:    "apple\nbanana\ncherry\n",
			expected: "cherry\nbanana\napple\n",
		},
		{
			name:     "Numeric in-memory",
			cfg:      config.Options{Numeric: true},
			ctx:      context.Background(),
			input:    "10\n2\n1\n",
			expected: "1\n2\n10\n",
		},
		{
			name:     "Unique in-memory",
			cfg:      config.Options{Unique: true},
			ctx:      context.Background(),
			input:    "apple\nbanana\napple\n",
			expected: "apple\nbanana\n",
		},
		{
			name:     "Months in-memory",
			cfg:      config.Options{Months: true},
			ctx:      context.Background(),
			input:    "march 3rd\nJAN 1st\nFeb 2nd\n",
			expected: "JAN 1st\nFeb 2nd\nmarch 3rd\n",
		},
		{
			name:     "Human suffix in-memory",
			cfg:      config.Options{HumanSuffix: true},
			ctx:      context.Background(),
			input:    "2000T\n10\n1000000M\n",
			expected: "10\n1000000M\n2000T\n",
		},
		{
			name:     "Column numeric in-memory",
			cfg:      config.Options{ColumnNum: 2, Numeric: true},
			ctx:      context.Background(),
			input:    "ID_B\t100\nID_A\t20\nID_C\t5\n",
			expected: "ID_C\t5\nID_A\t20\nID_B\t100\n",
		},
		{
			name:     "Ignore leading blanks in-memory",
			cfg:      config.Options{TrailingBlanks: true},
			ctx:      context.Background(),
			input:    "  b\na\n",
			expected: "a\n  b\n",
		},
		{
			name:    "Context cancelled in-memory",
			cfg:     config.Options{},
			ctx:     cancelledCtx,
			input:   "b\na\n",
			wantErr: true,
		},

		// --- external path (forced by small threshold) ---
		{
			name:      "Alphabetical external",
			cfg:       config.Options{},
			ctx:       context.Background(),
			input:     "banana\napple\ncherry\n",
			expected:  "apple\nbanana\ncherry\n",
			threshold: smallThreshold,
		},
		{
			name:      "Reverse external",
			cfg:       config.Options{Reverse: true},
			ctx:       context.Background(),
			input:     "apple\nbanana\ncherry\n",
			expected:  "cherry\nbanana\napple\n",
			threshold: smallThreshold,
		},
		{
			name:      "Numeric external",
			cfg:       config.Options{Numeric: true},
			ctx:       context.Background(),
			input:     "10\n2\n1\n",
			expected:  "1\n2\n10\n",
			threshold: smallThreshold,
		},
		{
			name:      "Unique external dedup during merge",
			cfg:       config.Options{Unique: true},
			ctx:       context.Background(),
			input:     "apple\nbanana\napple\ncherry\nbanana\n",
			expected:  "apple\nbanana\ncherry\n",
			threshold: smallThreshold,
		},
		{
			name:      "Months external",
			cfg:       config.Options{Months: true},
			ctx:       context.Background(),
			input:     "march 3rd\nJAN 1st\nFeb 2nd\n",
			expected:  "JAN 1st\nFeb 2nd\nmarch 3rd\n",
			threshold: smallThreshold,
		},
		{
			name:      "Ignore leading blanks external",
			cfg:       config.Options{TrailingBlanks: true},
			ctx:       context.Background(),
			input:     "  b\na\nc\nd\n",
			expected:  "a\n  b\nc\nd\n",
			threshold: smallThreshold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []sorter.Option
			if tt.threshold > 0 {
				opts = append(opts, sorter.WithThreshold(tt.threshold))
			}
			s := sorter.NewAutoSorter(&tt.cfg, opts...)
			reader := strings.NewReader(tt.input)
			writer := &bytes.Buffer{}

			err := s.Sort(tt.ctx, reader, writer)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Sort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if writer.String() != tt.expected {
				t.Errorf("got:\n%q\nwant:\n%q", writer.String(), tt.expected)
			}
		})
	}
}

func TestExternal_Sort(t *testing.T) {
	t.Run("Single run — all fits in one batch", func(t *testing.T) {
		// threshold is large enough to hold all input in one run, triggering merge of 1 file
		cfg := config.Options{}
		s := sorter.NewExternal(&cfg, sorter.WithThreshold(1000))
		input := "cherry\nbanana\napple\n"
		reader := strings.NewReader(input)
		writer := &bytes.Buffer{}

		if err := s.Sort(context.Background(), reader, writer); err != nil {
			t.Fatalf("Sort() error: %v", err)
		}
		want := "apple\nbanana\ncherry\n"
		if writer.String() != want {
			t.Errorf("got %q, want %q", writer.String(), want)
		}
	})

	t.Run("Multiple runs — k-way merge", func(t *testing.T) {
		cfg := config.Options{}
		s := sorter.NewExternal(&cfg, sorter.WithThreshold(smallThreshold))
		// Each line is ~6 bytes, threshold=50 → several runs
		input := "mango\nkiwi\nbanana\napple\ncherry\ndatе\nfig\ngrape\n"
		reader := strings.NewReader(input)
		writer := &bytes.Buffer{}

		if err := s.Sort(context.Background(), reader, writer); err != nil {
			t.Fatalf("Sort() error: %v", err)
		}
		lines := strings.Split(strings.TrimSuffix(writer.String(), "\n"), "\n")
		for i := 1; i < len(lines); i++ {
			if lines[i] < lines[i-1] {
				t.Errorf("output not sorted at position %d: %q < %q", i, lines[i], lines[i-1])
			}
		}
	})

	t.Run("Unique with multiple runs", func(t *testing.T) {
		cfg := config.Options{Unique: true}
		s := sorter.NewExternal(&cfg, sorter.WithThreshold(smallThreshold))
		input := "banana\napple\nbanana\napple\ncherry\nbanana\n"
		reader := strings.NewReader(input)
		writer := &bytes.Buffer{}

		if err := s.Sort(context.Background(), reader, writer); err != nil {
			t.Fatalf("Sort() error: %v", err)
		}
		want := "apple\nbanana\ncherry\n"
		if writer.String() != want {
			t.Errorf("got %q, want %q", writer.String(), want)
		}
	})

	t.Run("Temp files are removed after sort", func(t *testing.T) {
		// Capture temp files created during sort by checking before/after.
		before, err := os.ReadDir(os.TempDir())
		if err != nil {
			t.Fatalf("ReadDir temp: %v", err)
		}
		beforeNames := make(map[string]bool, len(before))
		for _, e := range before {
			beforeNames[e.Name()] = true
		}

		cfg := config.Options{}
		s := sorter.NewExternal(&cfg, sorter.WithThreshold(smallThreshold))
		input := strings.Repeat("cherry\nbanana\napple\n", 5)
		if err := s.Sort(context.Background(), strings.NewReader(input), &bytes.Buffer{}); err != nil {
			t.Fatalf("Sort() error: %v", err)
		}

		after, err := os.ReadDir(os.TempDir())
		if err != nil {
			t.Fatalf("ReadDir temp after: %v", err)
		}
		for _, e := range after {
			if strings.HasPrefix(e.Name(), "sort-util-run-") && !beforeNames[e.Name()] {
				t.Errorf("temp file not cleaned up: %s", e.Name())
			}
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		cfg := config.Options{}
		s := sorter.NewExternal(&cfg, sorter.WithThreshold(smallThreshold))
		err := s.Sort(ctx, strings.NewReader("b\na\n"), &bytes.Buffer{})
		if err == nil {
			t.Fatal("expected error on cancelled context, got nil")
		}
	})
}
