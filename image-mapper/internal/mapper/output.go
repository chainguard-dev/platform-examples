package mapper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Output writes mappings in a particular format
type Output func(w io.Writer, mappings []*Mapping) error

// NewOutput returns an output in the requested format
func NewOutput(format string) (Output, error) {
	switch strings.ToLower(format) {
	case "csv":
		return outputCSV, nil
	case "json":
		return outputJSON, nil
	case "text":
		return outputText, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s (supported: csv, json, text)", format)
	}
}

func outputCSV(w io.Writer, mappings []*Mapping) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	for _, m := range mappings {
		if err := writer.Write([]string{m.Image, fmt.Sprintf("%s", m.Results)}); err != nil {
			return fmt.Errorf("writing CSV record: %w", err)
		}
	}

	return nil
}

func outputJSON(w io.Writer, mappings []*Mapping) error {
	return json.NewEncoder(w).Encode(mappings)
}

func outputText(w io.Writer, mappings []*Mapping) error {
	for _, m := range mappings {
		for _, result := range m.Results {
			fmt.Fprintf(w, "%s -> %s\n", m.Image, result)
		}
		if len(m.Results) == 0 {
			fmt.Fprintf(w, "%s ->\n", m.Image)
		}
	}
	return nil
}
