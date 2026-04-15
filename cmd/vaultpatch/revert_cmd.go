package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/revert"
	"github.com/your-org/vaultpatch/internal/vault"
)

func init() {
	var (
		path    string
		version int
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "revert",
		Short: "Revert a secret to a previous KV v2 version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRevert(cmd.Context(), path, version, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Secret path to revert (required)")
	cmd.Flags().IntVar(&version, "version", 0, "Target version to revert to (required, >= 1)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview revert without writing")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("version")

	rootCmd.AddCommand(cmd)
}

func runRevert(ctx context.Context, path string, version int, dryRun bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	rv := revert.NewReverter(client)
	res := rv.Apply(ctx, revert.Options{
		Path:    path,
		Version: version,
		DryRun:  dryRun,
	})

	if res.Err != nil {
		return res.Err
	}

	if dryRun {
		fmt.Printf("[dry-run] would revert %q to version %d (%d keys)\n", path, version, len(res.Data))
	} else {
		fmt.Printf("reverted %q to version %d (%d keys)\n", path, version, len(res.Data))
	}
	return nil
}
