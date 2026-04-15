package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/youorg/vaultpatch/internal/checkpoint"
	"github.com/youorg/vaultpatch/internal/vault"
)

func init() {
	checkpointCmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Capture or inspect a labelled checkpoint of Vault secrets",
	}

	captureCmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a checkpoint for the given paths",
		RunE:  runCheckpointCapture,
	}
	captureCmd.Flags().String("label", "", "Label for this checkpoint (required)")
	captureCmd.Flags().StringSlice("paths", nil, "Vault paths to capture (comma-separated)")
	captureCmd.Flags().String("out", "checkpoint.json", "Output file path")
	_ = captureCmd.MarkFlagRequired("label")
	_ = captureCmd.MarkFlagRequired("paths")

	checkpointCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(checkpointCmd)
}

func runCheckpointCapture(cmd *cobra.Command, _ []string) error {
	label, _ := cmd.Flags().GetString("label")
	paths, _ := cmd.Flags().GetStringSlice("paths")
	out, _ := cmd.Flags().GetString("out")
	ns := getEnvOrDefault("VAULT_NAMESPACE", "")

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: ns,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	col := checkpoint.NewCollector(client, ns)
	ctx := context.Background()

	clean := make([]string, 0, len(paths))
	for _, p := range paths {
		clean = append(clean, strings.TrimSpace(p))
	}

	secrets, err := col.Collect(ctx, clean)
	if err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	cp := checkpoint.New(label, ns, secrets)
	if err := checkpoint.Save(cp, out); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "checkpoint %q saved to %s (%d paths)\n", label, out, len(secrets))
	return nil
}
