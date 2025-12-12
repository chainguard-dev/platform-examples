package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	ignoreTiers  []string
)

var rootCmd = &cobra.Command{
	Use:   "image-mapper",
	Short: "Map upstream image references to Chainguard images.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		output, err := mapper.NewOutput(outputFormat)
		if err != nil {
			return fmt.Errorf("constructing output: %w", err)
		}

		var opts []mapper.Option
		if len(ignoreTiers) > 0 {
			opts = append(opts, mapper.WithoutTiers(ignoreTiers))
		}
		m, err := mapper.NewMapper(ctx, opts...)
		if err != nil {
			return fmt.Errorf("creating mapper: %w", err)
		}

		it := mapper.NewArgsIterator(args)
		if args[0] == "-" {
			it = mapper.NewReaderIterator(os.Stdin)
		}

		mappings, err := m.MapAll(it)
		if err != nil {
			return fmt.Errorf("mapping images: %w", err)
		}

		return output(os.Stdout, mappings)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (csv, json, text, customer-yaml)")
	rootCmd.Flags().StringSliceVar(&ignoreTiers, "ignore-tiers", []string{}, "Ignore Chainguard repos of specific tiers (PREMIUM, APPLICATION, BASE, FIPS, AI)")
}

func Execute() error {
	return rootCmd.Execute()
}
