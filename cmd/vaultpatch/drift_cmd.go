package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/drift"
)

func init() {
	driftCmd := &cobra.Command{
		Use:   "drift <path> <baseline-file>",
		Short: "Detect drift between a baseline snapshot and live Vault secrets",
		Args:  cobra.ExactArgs(2),
		RunE:  runDrift,
	}
	driftCmd.Flags().Bool("json", false, "Output report as JSON")
	rootCmd.AddCommand(driftCmd)
}

func runDrift(cmd *cobra.Command, args []string) error {
	path := args[0]
	baselineFile := args[1]

	jsonOut, _ := cmd.Flags().GetBool("json")

	f, err := os.Open(baselineFile)
	if err != nil {
		return fmt.Errorf("drift: open baseline: %w", err)
	}
	defer f.Close()

	var baseline map[string]string
	if err := json.NewDecoder(f).Decode(&baseline); err != nil {
		return fmt.Errorf("drift: decode baseline: %w", err)
	}

	client, err := newVaultClient()
	if err != nil {
		return err
	}

	detector := drift.NewDetector(client)
	report, err := detector.Detect(context.Background(), path, baseline)
	if err != nil {
		return err
	}

	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(report)
	}

	fmt.Println(report.Summary())
	for _, d := range report.Deltas {
		switch d.Status {
		case "added":
			fmt.Printf("  + %-30s (current: %q)\n", d.Key, d.Current)
		case "removed":
			fmt.Printf("  - %-30s (baseline: %q)\n", d.Key, d.Baseline)
		case "changed":
			fmt.Printf("  ~ %-30s %q -> %q\n", d.Key, d.Baseline, d.Current)
		}
	}

	if report.HasDrift() {
		os.Exit(1)
	}
	return nil
}
