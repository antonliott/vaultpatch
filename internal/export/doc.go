// Package export provides multi-format export capabilities for Vault secret
// snapshots collected by the snapshot package.
//
// Supported formats:
//
//   - json    — pretty-printed JSON object mapping keys to values
//   - yaml    — YAML mapping suitable for use with Kubernetes or Helm
//   - dotenv  — shell-compatible KEY="value" pairs, sorted alphabetically
//
// Example usage:
//
//	exporter, err := export.NewExporter(export.FormatDotenv, os.Stdout)
//	if err != nil { ... }
//	if err := exporter.Write(secrets); err != nil { ... }
package export
