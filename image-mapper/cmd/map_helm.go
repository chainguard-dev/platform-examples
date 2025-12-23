package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/helm"
	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/spf13/cobra"
)

func MapHelmChartCommand() *cobra.Command {
	opts := struct {
		Repo         string
		ChartRepo    string
		ChartVersion string
	}{}
	cmd := &cobra.Command{
		Use:   "helm-chart",
		Short: "Extract image related values from a Helm chart and map them to Chainguard.",
		Example: `
  # Map a Helm chart. This requires that the Chart repo has been added with 'helm repo add' beforehand.
  image-mapper map helm-chart argocd/argo-cd
  
  # Override the repository in the mappings with your own mirror or proxy. For instance, cgr.dev/chainguard/<image> would become registry.internal/cgr/<image> in the output.
  image-mapper map helm-chart argocd/argo-cd --repository=registry.internal/cgr
  
  # Map a specific version of a Helm chart.
  image-mapper map helm-chart argocd/argo-cd --chart-version=9.0.0
  
  # Specify a remote Chart repostory. This means the repository doesn't need to be added with 'helm repo add'.
  image-mapper map helm-chart argo-cd --chart-repo=https://argoproj.github.io/argo-helm
  
  # Specify a specific version of a remote Chart.
  image-mapper map helm-chart argo-cd --chart-repo=https://argoproj.github.io/argo-helm --chart-version=9.0.0
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			chart := helm.ChartDescriptor{
				Name:       args[0],
				Repository: opts.ChartRepo,
				Version:    opts.ChartVersion,
			}
			output, err := helm.MapChart(ctx, chart, mapper.WithRepository(opts.Repo))
			if err != nil {
				return fmt.Errorf("mapping values: %w", err)
			}

			if _, err := os.Stdout.Write(output); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Repo, "repository", "cgr.dev/chainguard", "Modifies the repository URI in the mappings. For instance, registry.internal.dev/chainguard would result in registry.internal.dev/chainguard/<image> in the output.")
	cmd.Flags().StringVar(&opts.ChartRepo, "chart-repo", "", "The chart repository url to locate the requested chart.")
	cmd.Flags().StringVar(&opts.ChartVersion, "chart-version", "", "A version constraint for the chart version.")

	return cmd
}

func MapHelmValuesCommand() *cobra.Command {
	opts := struct {
		Repo string
	}{}
	cmd := &cobra.Command{
		Use:   "helm-values",
		Short: "Extract image related values from a Helm values file and map them to Chainguard.",
		Example: `
  # Map images in the values returned by 'helm show values'
  helm show values argocd/argo-cd | image-mapper map helm-values -
  
  # Map images in a values file on disk.
  helm show values argocd/argo-cd > values.yaml
  image-mapper map helm-values values.yaml
  
  # Override the repository in the mappings with your own mirror or proxy. For instance, cgr.dev/chainguard/<image> would become registry.internal/cgr/<image> in the output.
  image-mapper map helm-values values.yaml --repository=registry.internal/cgr
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

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

			output, err := helm.MapValues(ctx, input, mapper.WithRepository(opts.Repo))
			if err != nil {
				return fmt.Errorf("mapping values: %w", err)
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
