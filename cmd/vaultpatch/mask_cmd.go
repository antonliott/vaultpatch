package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/youorg/vaultpatch/internal/mask"
	"github.com/youorg/vaultpatch/internal/vault"
)

var maskCmd = &cobra.Command{
	Use:   "mask <path>",
	Short: "Print secrets at a KV path with sensitive values redacted",
	Args:  cobra.ExactArgs(1),
	RunE:  runMask,
}

var maskPatterns []string

func init() {
	maskCmd.Flags().StringSliceVar(&maskPatterns, "pattern", mask.DefaultPatterns,
		"Key patterns to redact (case-insensitive substrings)")
	rootCmd.AddCommand(maskCmd)
}

func runMask(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	secrets, err := vault.ReadSecrets(cmd.Context(), client, path)
	if err != nil {
		return fmt.Errorf("read secrets: %w", err)
	}

	m, err := mask.New(maskPatterns)
	if err != nil {
		return fmt.Errorf("build masker: %w", err)
	}

	safe := m.Apply(secrets)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(safe); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	return nil
}
