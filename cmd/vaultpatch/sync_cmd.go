package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/sync"
	"github.com/example/vaultpatch/internal/vault"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronise secrets from a source namespace to a destination namespace",
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().String("src-namespace", "", "Source Vault namespace (required)")
	syncCmd.Flags().String("dst-namespace", "", "Destination Vault namespace (required)")
	syncCmd.Flags().String("mount", "secret", "KV mount to sync")
	syncCmd.Flags().String("prefix", "", "Key prefix to sync")
	syncCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	_ = syncCmd.MarkFlagRequired("src-namespace")
	_ = syncCmd.MarkFlagRequired("dst-namespace")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, _ []string) error {
	srcNS, _ := cmd.Flags().GetString("src-namespace")
	dstNS, _ := cmd.Flags().GetString("dst-namespace")
	mount, _ := cmd.Flags().GetString("mount")
	prefix, _ := cmd.Flags().GetString("prefix")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	addr := getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200")
	token := os.Getenv("VAULT_TOKEN")

	srcClient, err := vault.NewClient(vault.Config{Address: addr, Token: token, Namespace: srcNS})
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}

	dstClient, err := vault.NewClient(vault.Config{Address: addr, Token: token, Namespace: dstNS})
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	syncer := sync.NewSyncer(srcClient, dstClient)
	res := syncer.Apply(context.Background(), sync.Options{
		Mount:  mount,
		Prefix: prefix,
		DryRun: dryRun,
	})

	log.Println(res.Summary())

	if res.HasErrors() {
		for _, e := range res.Errors {
			log.Printf("  error: %v", e)
		}
		return fmt.Errorf("sync completed with %d error(s)", len(res.Errors))
	}
	return nil
}
