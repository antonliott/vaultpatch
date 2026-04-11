// Package export provides functionality to export Vault secret snapshots
// to various output formats (JSON, YAML, dotenv).
package export

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents a supported export format.
type Format string

const (
	FormatJSON   Format = "json"
	FormatYAML   Format = "yaml"
	FormatDotenv Format = "dotenv"
)

// Exporter writes secret data to an io.Writer in a specified format.
type Exporter struct {
	format Format
	writer io.Writer
}

// NewExporter creates a new Exporter for the given format and writer.
func NewExporter(format Format, w io.Writer) (*Exporter, error) {
	switch format {
	case FormatJSON, FormatYAML, FormatDotenv:
		return &Exporter{format: format, writer: w}, nil
	default:
		return nil, fmt.Errorf("unsupported export format: %q", format)
	}
}

// Write serializes the provided secrets map to the configured format.
func (e *Exporter) Write(secrets map[string]string) error {
	switch e.format {
	case FormatJSON:
		return writeJSON(e.writer, secrets)
	case FormatYAML:
		return writeYAML(e.writer, secrets)
	case FormatDotenv:
		return writeDotenv(e.writer, secrets)
	default:
		return fmt.Errorf("unsupported format: %q", e.format)
	}
}

func writeJSON(w io.Writer, secrets map[string]string) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(secrets)
}

func writeYAML(w io.Writer, secrets map[string]string) error {
	return yaml.NewEncoder(w).Encode(secrets)
}

func writeDotenv(w io.Writer, secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := strings.ReplaceAll(secrets[k], `"`, `\"`)
		sb.WriteString(fmt.Sprintf("%s=\"%s\"\n", k, v))
	}
	_, err := fmt.Fprint(w, sb.String())
	return err
}
