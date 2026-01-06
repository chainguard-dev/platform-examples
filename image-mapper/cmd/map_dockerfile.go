package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/dockerfile"
	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/spf13/cobra"
)

func MapDockerfileCommand() *cobra.Command {
	opts := struct {
		Repo string
	}{}
	cmd := &cobra.Command{
		Use:   "dockerfile",
		Short: "Map image references in a Dockerfile to their Chainguard equivalents.",
		Example: `
# Map a Dockerfile
image-mapper map dockerfile Dockerfile

# Map a Dockerfile from stdin
cat Dockerfile | image-mapper map dockerfile -

# Override the repository in the mappings with your own mirror or proxy. For instance, cgr.dev/chainguard/<image> would become registry.internal/cgr/<image> in the output.
image-mapper map dockerfile Dockerfile --repository=registry.internal/cgr
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				input []byte
				err   error
			)
			switch args[0] {
			case "-":
				input, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
			default:
				input, err = os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("reading file: %s: %w", args[0], err)
				}
			}

			output, err := dockerfile.Map(cmd.Context(), input, mapper.WithRepository(opts.Repo))
			if err != nil {
				return fmt.Errorf("mapping dockerfile: %w", err)
			}

			if _, err := os.Stdout.Write(output); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Repo, "repository", "cgr.dev/chainguard", "Modifies the repository URI in the mappings. For instance, registry.internal.dev/chainguard would result in registry.internal.dev/chainguard/<image> in the output.")

	return cmd
}
