package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/archive"
	"github.com/your-org/vaultpatch/internal/vault"
)

var (
	archiveDryRun bool
	archiveRoot   string
)

func init() {
	archiveCmd := &cobra.Command{
		Use:   "archive [path...]",
		Short: "Soft-delete secrets by moving them to an archive prefix",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runArchive,
	}
	archiveCmd.Flags().BoolVar(&archiveDryRun, "dry-run", false, "Preview changes without writing")
	archiveCmd.Flags().StringVar(&archiveRoot, "archive-root", "secret/archive", "Destination prefix for archived secrets")
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	a := archive.NewArchiver(client, client, archiveRoot, archiveDryRun)
	res, err := a.Apply(cmd.Context(), args)
	if err != nil {
		return err
	}

	if archiveDryRun {
		fmt.Println("[dry-run] would archive:")
	} else {
		fmt.Println("archived:")
	}
	for _, p := range res.Archived {
		fmt.Printf("  %s -> %s/%s\n", p, archiveRoot, p)
	}
	for _, p := range res.Skipped {
		fmt.Printf("  skipped (not found): %s\n", p)
	}
	if res.HasErrors() {
		msgs := make([]string, len(res.Errors))
		for i, e := range res.Errors {
			msgs[i] = e.Error()
		}
		return fmt.Errorf("archive completed with errors:\n  %s", strings.Join(msgs, "\n  "))
	}
	return nil
}
