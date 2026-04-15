package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/mirror"
	"github.com/your-org/vaultpatch/internal/vault"
)

var (
	mirrorSrc       string
	mirrorDst       string
	mirrorOverwrite bool
	mirrorDryRun    bool
)

func init() {
	mirrorCmd := &cobra.Command{
		Use:   "mirror",
		Short: "Mirror secrets from one path prefix to another",
		RunE:  runMirror,
	}

	mirrorCmd.Flags().StringVar(&mirrorSrc, "src", "", "Source path prefix (required)")
	mirrorCmd.Flags().StringVar(&mirrorDst, "dst", "", "Destination path prefix (required)")
	mirrorCmd.Flags().BoolVar(&mirrorOverwrite, "overwrite", false, "Overwrite existing secrets at destination")
	mirrorCmd.Flags().BoolVar(&mirrorDryRun, "dry-run", false, "Preview changes without writing")
	_ = mirrorCmd.MarkFlagRequired("src")
	_ = mirrorCmd.MarkFlagRequired("dst")

	rootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, _ []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	secrets := vault.NewSecrets(client)
	m := mirror.NewMirrorer(secrets, secrets, mirrorOverwrite, mirrorDryRun)

	ctx := context.Background()
	res, err := m.Apply(ctx, mirrorSrc, mirrorDst)
	if err != nil {
		return err
	}

	cmd.Println(res.Summary())
	for _, p := range res.Mirrored {
		cmd.Printf("  mirrored: %s\n", p)
	}
	for _, p := range res.Skipped {
		cmd.Printf("  skipped:  %s\n", p)
	}
	for _, e := range res.Errors {
		cmd.Printf("  error:    %v\n", e)
	}

	if res.HasErrors() {
		return fmt.Errorf("mirror completed with %d error(s)", len(res.Errors))
	}
	return nil
}
