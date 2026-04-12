package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/rollback"
	"github.com/your-org/vaultpatch/internal/snapshot"
	"github.com/your-org/vaultpatch/internal/vault"
)

var (
	rollbackSnapshotFile string
	rollbackDryRun       bool
)

func init() {
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Revert secrets to a previously captured snapshot",
		RunE:  runRollback,
	}
	rollbackCmd.Flags().StringVar(&rollbackSnapshotFile, "snapshot", "", "Path to snapshot file (required)")
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "Preview changes without writing")
	_ = rollbackCmd.MarkFlagRequired("snapshot")
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, _ []string) error {
	snap, err := snapshot.Load(rollbackSnapshotFile)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: getEnvOrDefault("VAULT_NAMESPACE", snap.Namespace),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	rb := rollback.NewRollbacker(client, rollbackDryRun)
	res, err := rb.Apply(cmd.Context(), snap)
	if err != nil {
		return err
	}

	log.Println(res.Summary())

	for _, pe := range res.Errors {
		log.Printf("  error %s: %v", pe.Path, pe.Err)
	}

	if res.HasErrors() {
		return fmt.Errorf("rollback completed with %d error(s)", len(res.Errors))
	}
	return nil
}
