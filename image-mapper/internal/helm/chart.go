package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/yamlhelpers"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// ChartDescriptor describes a chart
type ChartDescriptor struct {
	Name       string
	Repository string
	Version    string
}

// MapChart extracts image related values from a Helm chart and maps them to
// Chainguard
func MapChart(ctx context.Context, chart ChartDescriptor, opts ...mapper.Option) ([]byte, error) {
	// Create a temporary directory where we'll untar the chart
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("creating temporary directory: %w", err)
	}
	defer os.RemoveAll(dir)

	// Pull the helm chart down to the temp dir
	if err := helmPull(ctx, chart, dir); err != nil {
		return nil, fmt.Errorf("pulling chart: %w", err)
	}

	// Construct a mapper
	m, err := NewMapper(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("constructing mapper: %w", err)
	}

	// Map the images in the chart
	return mapChart(m, dir)
}

// mapChart extracts image related values from the chart and maps them to
// Chainguard
func mapChart(m mapper.Mapper, chartPath string) ([]byte, error) {
	// Collect all the values.yaml files in the Chart and its subcharts
	var valuesFiles []string
	if err := filepath.WalkDir(chartPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		if _, err := os.Stat(filepath.Join(path, "Chart.yaml")); err != nil {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, "values.yaml")); err != nil {
			return nil
		}

		valuesFiles = append(valuesFiles, filepath.Join(path, "values.yaml"))

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking chart directory: %w", err)
	}

	// We'll write modified nodes to this node
	outputNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	// Iterate backwards over the collected values files so that we map the
	// child values before the parents, prefering any overrides configured
	// in the parent values.
	for i := len(valuesFiles) - 1; i >= 0; i-- {
		path := valuesFiles[i]

		yamlPath := buildPath(strings.TrimPrefix(path, chartPath))

		inputNode, err := readValuesFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading values file: %s: %w", path, err)
		}

		if err := yamlhelpers.WalkNode(inputNode, mapNode(m, yamlPath, outputNode)); err != nil {
			return nil, err
		}

	}

	// Marshal the modified nodes to a new document
	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{outputNode},
	}
	output, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshalling output document: %w", err)
	}

	return output, nil
}

// readValuesFile reads a values file from disk and returns it as a *yaml.Node
func readValuesFile(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}

	// Return root mapping node (skip document wrapper)
	if len(doc.Content) > 0 {
		return doc.Content[0], nil
	}

	return &yaml.Node{Kind: yaml.MappingNode}, nil
}

// buildPath infers the appropriate nesting in the yaml structure based on the
// path to a values file.
//
// For instance, values in charts/grafana/values.yaml would be nested under
// "grafana".
//
// And values in charts/grafana/charts/redis/values.yaml would be nested under
// "grafana.redis".
func buildPath(path string) []string {
	parent := []string{}
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for i, part := range parts {
		if part != "charts" {
			continue
		}

		parent = append(parent, parts[i+1])
	}

	return parent
}

// helmPull pulls a remote chart and extracts it to the specified directory
func helmPull(ctx context.Context, chart ChartDescriptor, dir string) error {
	client := action.NewPullWithOpts(action.WithConfig(&action.Configuration{}))
	client.Settings = cli.New()
	client.DestDir = dir
	client.Untar = true

	if chart.Version != "" {
		client.Version = chart.Version
	}
	if chart.Repository != "" {
		client.RepoURL = chart.Repository
	}

	_, err := client.Run(chart.Name)
	if err != nil {
		return fmt.Errorf("pulling chart: %w", err)
	}

	return nil
}
