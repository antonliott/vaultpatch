package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/lock"
	"github.com/example/vaultpatch/internal/vault"
)

var lockCmd = &cobra.Command{
	Use:   "lock <secret-path>",
	Short: "Acquire or release a distributed lock on a Vault secret path",
	Args:  cobra.ExactArgs(1),
	RunE:  runLock,
}

var (
	lockRelease bool
	lockOwner   string
	lockTTL     time.Duration
	lockMount   string
)

func init() {
	lockCmd.Flags().BoolVar(&lockRelease, "release", false, "Release an existing lock instead of acquiring")
	lockCmd.Flags().StringVar(&lockOwner, "owner", "", "Owner identifier for the lock (default: hostname)")
	lockCmd.Flags().DurationVar(&lockTTL, "ttl", 60*time.Second, "Lock TTL duration")
	lockCmd.Flags().StringVar(&lockMount, "mount", "secret", "KV mount used to store lock metadata")
	rootCmd.AddCommand(lockCmd)
}

func runLock(cmd *cobra.Command, args []string) error {
	secretPath := args[0]

	owner := lockOwner
	if owner == "" {
		h, err := os.Hostname()
		if err != nil {
			h = "unknown"
		}
		owner = h
	}

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	lockPath := fmt.Sprintf("%s/locks/%s", lockMount, secretPath)
	l := lock.NewLock(client, lockPath, owner, lockTTL)
	ctx := context.Background()

	if lockRelease {
		if err := l.Release(ctx); err != nil {
			return fmt.Errorf("release lock: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "lock released: %s\n", lockPath)
		return nil
	}

	if err := l.Acquire(ctx); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "lock acquired: %s (owner=%s ttl=%s)\n", lockPath, owner, lockTTL)
	return nil
}
