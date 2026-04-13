package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	vaultcopy "github.com/example/vaultpatch/internal/copy"
	"github.com/example/vaultpatch/internal/vault"
)

var copyCmd = &cobra.Command{
	Use:   "copy <src-path> <dst-path>",
	Short: "Copy secrets from one Vault path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runCopy,
}

func init() {
	copyCmd.Flags().Bool("dry-run", false, "Preview the copy without writing to Vault")
	copyCmd.Flags().String("src-namespace", "", "Source Vault namespace (overrides VAULT_NAMESPACE)")
	copyCmd.Flags().String("dst-namespace", "", "Destination Vault namespace")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	srcNS, _ := cmd.Flags().GetString("src-namespace")
	dstNS, _ := cmd.Flags().GetString("dst-namespace")

	srcClient, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: getEnvOrDefault("VAULT_NAMESPACE", srcNS),
	})
	if err != nil {
		return fmt.Errorf("source vault client: %w", err)
	}

	dstNSResolved := dstNS
	if dstNSResolved == "" {
		dstNSResolved = getEnvOrDefault("VAULT_NAMESPACE", "")
	}

	dstClient, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: dstNSResolved,
	})
	if err != nil {
		return fmt.Errorf("destination vault client: %w", err)
	}

	copier := vaultcopy.NewCopier(srcClient, dstClient, dryRun)
	result := copier.Apply(context.Background(), srcPath, dstPath)

	fmt.Fprintln(cmd.OutOrStdout(), result.Summary())

	if result.Err != nil {
		return result.Err
	}
	return nil
}
