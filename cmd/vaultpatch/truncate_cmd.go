package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/truncate"
	"github.com/your-org/vaultpatch/internal/vault"
)

func init() {
	var (
		path   string
		maxLen int
		suffix string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "truncate",
		Short: "Truncate secret values that exceed a maximum length",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTruncate(path, maxLen, suffix, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault KV path to scan (required)")
	cmd.Flags().IntVar(&maxLen, "max-len", 64, "Maximum allowed value length")
	cmd.Flags().StringVar(&suffix, "suffix", "...", "Suffix appended to truncated values")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runTruncate(path string, maxLen int, suffix string, dryRun bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	tr := truncate.NewTruncator(client, truncate.Options{
		MaxLen: maxLen,
		Suffix: suffix,
		DryRun: dryRun,
	})

	results, err := tr.Apply(path)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("no values exceeded max length")
		return nil
	}

	for _, r := range results {
		fmt.Println(r)
	}
	fmt.Printf("\n%d value(s) truncated", len(results))
	if dryRun {
		fmt.Print(" (dry-run)")
	}
	fmt.Println()
	return nil
}
