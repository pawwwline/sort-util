// Package cmd provides the command-line interface implementation using Cobra.
package cmd

import (
	"fmt"
	"io"
	"os"

	"sort-util/internal/app"
	"sort-util/internal/config"
	"sort-util/internal/sorter"

	"github.com/spf13/cobra"
)

var cfg config.Options

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sort-util",
	Short: "Sort lines of text files",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var reader io.Reader = os.Stdin
		if len(args) > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("open input file: %w", err)
			}

			defer func() { _ = file.Close() }()

			reader = file
		}

		engine := sorter.NewInMemory(cfg)
		application := app.New(engine)

		return application.Run(cmd.Context(), reader, os.Stdout)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&cfg.Numeric, "numeric", "n", false, "compare according to string numerical value")
	rootCmd.Flags().BoolVarP(&cfg.Unique, "unique", "u", false, "output only unique values")
	rootCmd.Flags().BoolVarP(&cfg.Reverse, "reverse", "r", false, "reverse sort order")
	rootCmd.Flags().BoolVarP(&cfg.TrailingBlanks, "blanks", "b", false, "remove trailing and leading blanks")
}
