package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/digest"
	"github.com/your-org/vaultpatch/internal/vault"
)

var digestCmd = &cobra.Command{
	Use:   "digest <mount-path>",
	Short: "Compute SHA-256 digests for all secrets under a mount path",
	Args:  cobra.ExactArgs(1),
	RunE:  runDigest,
}

var digestJSONOutput bool

func init() {
	rootCmd.AddCommand(digestCmd)
	digestCmd.Flags().BoolVar(&digestJSONOutput, "json", false, "output results as JSON")
}

func runDigest(cmd *cobra.Command, args []string) error {
	mountPath := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	result, err := digest.Compute(client, mountPath)
	if err != nil {
		return fmt.Errorf("digest: %w", err)
	}

	if digestJSONOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if len(result) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no secrets found")
		return nil
	}

	for path, entry := range result {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", entry.Digest[:12], path)
	}
	return nil
}
