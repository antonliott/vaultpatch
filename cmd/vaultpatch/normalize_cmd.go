package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/normalize"
	"github.com/example/vaultpatch/internal/vault"
)

func init() {
	var (
		path   string
		style  string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize secret key casing in a Vault path",
		Long: `Reads all keys at the given KV v2 path and rewrites them
using the chosen style (upper, lower, snake, kebab).

Use --dry-run to preview changes without writing to Vault.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNormalize(path, normalize.Style(style), dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "KV v2 secret path (required)")
	cmd.Flags().StringVar(&style, "style", "upper", "Key style: upper|lower|snake|kebab")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runNormalize(path string, style normalize.Style, dryRun bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	secrets, err := client.Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	norm, err := normalize.New(style)
	if err != nil {
		return err
	}

	normalized := norm.Apply(secrets)

	if dryRun {
		fmt.Printf("[dry-run] would write %d key(s) to %s\n", len(normalized), path)
		for k := range normalized {
			fmt.Printf("  %s\n", k)
		}
		return nil
	}

	if err := client.Write(path, normalized); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	fmt.Printf("normalized %d key(s) at %s using style %q\n", len(normalized), path, style)
	return nil
}
