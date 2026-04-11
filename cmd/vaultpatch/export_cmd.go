package main

import (
	"fmt"
	"os"

	"github.com/your-org/vaultpatch/internal/export"
	"github.com/your-org/vaultpatch/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportInput  string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a snapshot to a specified format (json, yaml, dotenv)",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json, yaml, dotenv")
	exportCmd.Flags().StringVarP(&exportInput, "input", "i", "", "Path to snapshot file (required)")
	_ = exportCmd.MarkFlagRequired("input")
}

func runExport(cmd *cobra.Command, _ []string) error {
	snap, err := snapshot.Load(exportInput)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	exporter, err := export.NewExporter(export.Format(exportFormat), os.Stdout)
	if err != nil {
		return fmt.Errorf("creating exporter: %w", err)
	}

	if err := exporter.Write(snap.Secrets); err != nil {
		return fmt.Errorf("writing export: %w", err)
	}

	return nil
}
