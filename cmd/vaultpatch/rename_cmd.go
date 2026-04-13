package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/rename"
	"github.com/your-org/vaultpatch/internal/vault"
)

var renameDryRun bool

var renameCmd = &cobra.Command{
	Use:   "rename <path> <old-key> <new-key>",
	Short: "Rename a key within a Vault KV secret",
	Args:  cobra.ExactArgs(3),
	RunE:  runRename,
}

func init() {
	renameCmd.Flags().BoolVar(&renameDryRun, "dry-run", false, "preview changes without writing to Vault")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	path, oldKey, newKey := args[0], args[1], args[2]

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	r := rename.NewRenamer(client)
	res := r.Apply(context.Background(), path, oldKey, newKey, renameDryRun)

	switch {
	case res.Err != nil:
		return res.Err
	case res.Skipped:
		log.Printf("skip: key %q not found in %s", oldKey, path)
	case res.DryRun:
		log.Printf("dry-run: would rename %q → %q in %s", oldKey, newKey, path)
	default:
		log.Printf("renamed %q → %q in %s", oldKey, newKey, path)
	}
	return nil
}
