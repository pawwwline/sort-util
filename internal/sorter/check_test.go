package sorter_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"sort-util/internal/config"
	"sort-util/internal/sorter"
)

func TestChecker_CheckSorted(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name    string
		cfg     config.Options
		ctx     context.Context
		input   string
		wantErr bool
	}{
		{
			name:    "Empty input",
			cfg:     config.Options{},
			ctx:     context.Background(),
			input:   "",
			wantErr: false,
		},
		{
			name:    "Single line input",
			cfg:     config.Options{},
			ctx:     context.Background(),
			input:   "string\n",
			wantErr: false,
		},
		{
			name:    "Default alphabetical sort sorted",
			cfg:     config.Options{},
			ctx:     context.Background(),
			input:   "apple\nbanana\ncherry\n",
			wantErr: false,
		},
		{
			name:    "Default alphabetical sort not sorted",
			cfg:     config.Options{},
			ctx:     context.Background(),
			input:   "banana\napple\n",
			wantErr: true,
		},
		{
			name:    "Reverse alphabetical sort error",
			cfg:     config.Options{Reverse: true},
			ctx:     context.Background(),
			input:   "apple\nbanana\ncherry\n",
			wantErr: true,
		},
		{
			name:    "Reverse alphabetical sort",
			cfg:     config.Options{Reverse: true},
			ctx:     context.Background(),
			input:   "cherry\nbanana\napple\n",
			wantErr: false,
		},
		{
			name:    "Numeric sort",
			cfg:     config.Options{Numeric: true},
			ctx:     context.Background(),
			input:   "1\n2\n10\n",
			wantErr: false,
		},
		{
			name:    "Numeric sort and reverse error",
			cfg:     config.Options{Numeric: true, Reverse: true},
			ctx:     context.Background(),
			input:   "1\n10\n20\n",
			wantErr: true,
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
			c := sorter.NewChecker(tt.cfg)
			reader := strings.NewReader(tt.input)

			err := c.CheckSorted(tt.ctx, reader)

			switch {
			case tt.wantErr && err == nil:
				t.Fatalf("expected error, got nil")

			case !tt.wantErr && err != nil:
				t.Fatalf("unexpected error: %v", err)

			case tt.wantErr && tt.ctx.Err() != nil:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected context.Canceled, got %v", err)
				}
			}
		})
	}
}
