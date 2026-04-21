package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/example/vaultpatch/internal/baseline"
	"github.com/spf13/cobra"
)

func init() {
	captureCmd := &cobra.Command{
		Use:   "baseline-capture",
		Short: "Capture a baseline snapshot of secrets at a path",
		RunE:  runBaselineCapture,
	}
	captureCmd.Flags().String("path", "", "Vault secret path to capture (required)")
	captureCmd.Flags().String("out", "baseline.json", "Output file for the baseline")
	_ = captureCmd.MarkFlagRequired("path")

	driftCmd := &cobra.Command{
		Use:   "baseline-drift",
		Short: "Compare current secrets against a saved baseline",
		RunE:  runBaselineDrift,
	}
	driftCmd.Flags().String("path", "", "Vault secret path to compare (required)")
	driftCmd.Flags().String("baseline", "baseline.json", "Baseline file to compare against")
	_ = driftCmd.MarkFlagRequired("path")

	rootCmd.AddCommand(captureCmd, driftCmd)
}

func runBaselineCapture(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("path")
	out, _ := cmd.Flags().GetString("out")

	client, err := newVaultClient(cmd)
	if err != nil {
		return err
	}

	ns := getEnvOrDefault("VAULT_NAMESPACE", "")
	b, err := baseline.Capture(context.Background(), client, ns, path)
	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}
	if err := baseline.Save(b, out); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "baseline captured to %s (%d secrets)\n", out, len(b.Secrets))
	return nil
}

func runBaselineDrift(cmd *cobra.Command, _ []string) error {
	path, _ := cmd.Flags().GetString("path")
	baselineFile, _ := cmd.Flags().GetString("baseline")

	b, err := baseline.Load(baselineFile)
	if err != nil {
		return fmt.Errorf("load baseline: %w", err)
	}

	client, err := newVaultClient(cmd)
	if err != nil {
		return err
	}

	deltas, err := baseline.Compare(context.Background(), client, b, path)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	if len(deltas) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no drift detected")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tKEY\tSTATUS")
	for _, d := range deltas {
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.Path, d.Key, d.Status)
	}
	w.Flush()

	if len(deltas) > 0 {
		os.Exit(1)
	}
	return nil
}
