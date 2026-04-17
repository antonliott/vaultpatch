package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/filter"
	"github.com/your-org/vaultpatch/internal/vault"
)

var filterCmd = &cobra.Command{
	Use:   "filter <path>",
	Short: "Filter secret keys/values by pattern and print matching entries",
	Args:  cobra.ExactArgs(1),
	RunE:  runFilter,
}

var filterPatterns []string

func init() {
	filterCmd.Flags().StringArrayVarP(&filterPatterns, "match", "m", nil,
		"Pattern to match (key or key=value regex). Repeatable.")
	rootCmd.AddCommand(filterCmd)
}

func runFilter(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	secrets, err := client.Read(cmd.Context(), path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	f, err := filter.New(filterPatterns)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	matched := f.Apply(secrets)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(matched); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	return nil
}
