package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/immute"
	"github.com/your-org/vaultpatch/internal/vault"
)

func init() {
	var dryRun bool
	var paths []string

	cmd := &cobra.Command{
		Use:   "immute",
		Short: "Mark secrets as immutable to prevent further writes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImmute(paths, dryRun)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().StringSliceVar(&paths, "path", nil, "Secret paths to mark immutable (comma-separated)")
	_ = cmd.MarkFlagRequired("path")
	rootCmd.AddCommand(cmd)
}

func runImmute(paths []string, dryRun bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	rw := vault.NewSecretsRW(client)
	locker := immute.NewLocker(rw, dryRun)
	res, err := locker.Apply(context.Background(), paths)
	if err != nil {
		return err
	}

	fmt.Println(res.Summary())
	for _, p := range res.Marked {
		fmt.Printf("  marked:  %s\n", p)
	}
	for _, p := range res.Skipped {
		fmt.Printf("  skipped: %s\n", p)
	}
	if res.HasErrors() {
		fmt.Fprintln(os.Stderr, "errors:")
		for _, e := range res.Errors {
			fmt.Fprintf(os.Stderr, "  %s\n", strings.TrimSpace(e))
		}
		return fmt.Errorf("immute completed with %d error(s)", len(res.Errors))
	}
	return nil
}
