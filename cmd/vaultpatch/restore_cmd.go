package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/restore"
	"github.com/example/vaultpatch/internal/snapshot"
)

var (
	restoreFile   string
	restoreDryRun bool
)

func init() {
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore secrets from a snapshot file into Vault",
		RunE:  runRestore,
	}
	restoreCmd.Flags().StringVarP(&restoreFile, "file", "f", "", "Path to snapshot file (required)")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Preview restore without writing to Vault")
	_ = restoreCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, _ []string) error {
	snap, err := snapshot.Load(restoreFile)
	if err != nil {
		return fmt.Errorf("loading snapshot: %w", err)
	}

	client, err := newVaultClient()
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	r := restore.NewRestorer(client, restoreDryRun)
	result, err := r.Apply(context.Background(), snap)
	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Fprintln(os.Stdout, result.Summary())

	for _, e := range result.Errors {
		log.Printf("ERROR: %v", e)
	}

	if result.HasErrors() {
		return fmt.Errorf("%d secret(s) failed to restore", len(result.Errors))
	}
	return nil
}
