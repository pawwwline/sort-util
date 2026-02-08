package sorter_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sorter.NewInMemory(tt.cfg)
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
