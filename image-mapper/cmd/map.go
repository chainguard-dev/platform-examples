package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(
		MapCommand(),
	)
}

func MapCommand() *cobra.Command {
	opts := struct {
		OutputFormat     string
		IgnoreTiers      []string
		IgnoreIamguarded bool
		Repo             string
	}{}
	cmd := &cobra.Command{
		Use:   "map",
		Short: "Map upstream image references to Chainguard images.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			output, err := mapper.NewOutput(opts.OutputFormat)
			if err != nil {
				return fmt.Errorf("constructing output: %w", err)
			}

			var ignoreFns []mapper.IgnoreFn
			if len(opts.IgnoreTiers) > 0 {
				ignoreFns = append(ignoreFns, mapper.IgnoreTiers(opts.IgnoreTiers))
			}
			if opts.IgnoreIamguarded {
				ignoreFns = append(ignoreFns, mapper.IgnoreIamguarded())
			}
			m, err := mapper.NewMapper(ctx, mapper.WithRepository(opts.Repo), mapper.WithIgnoreFns(ignoreFns...))
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

	rootCmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "Output format (csv, json, text, customer-yaml)")
	rootCmd.Flags().StringSliceVar(&opts.IgnoreTiers, "ignore-tiers", []string{}, "Ignore Chainguard repos of specific tiers (PREMIUM, APPLICATION, BASE, FIPS, AI)")
	rootCmd.Flags().BoolVar(&opts.IgnoreIamguarded, "ignore-iamguarded", false, "Ignore iamguarded images")
	rootCmd.Flags().StringVar(&opts.Repo, "repository", "cgr.dev/chainguard", "Modifies the repository URI in the mappings. For instance, registry.internal.dev/chainguard would result in registry.internal.dev/chainguard/<image> in the output.")

	cmd.AddCommand(
		MapHelmChartCommand(),
		MapHelmValuesCommand(),
	)

	return cmd
}
