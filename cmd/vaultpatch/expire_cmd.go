package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/expire"
	"github.com/your-org/vaultpatch/internal/vault"
)

var (
	expirePath    string
	expireMetaKey string
	expireWarn    time.Duration
)

func init() {
	cmd := &cobra.Command{
		Use:   "expire",
		Short: "Scan secrets for expiry metadata and report expired or expiring keys",
		RunE:  runExpire,
	}
	cmd.Flags().StringVar(&expirePath, "path", "", "KV path to scan (required)")
	cmd.Flags().StringVar(&expireMetaKey, "meta-key", "expires_at", "Secret field holding RFC3339 expiry timestamp")
	cmd.Flags().DurationVar(&expireWarn, "warn-within", 7*24*time.Hour, "Warn if secret expires within this duration")
	_ = cmd.MarkFlagRequired("path")
	rootCmd.AddCommand(cmd)
}

func runExpire(cmd *cobra.Command, _ []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return err
	}

	checker := expire.NewChecker(client, expireMetaKey, expireWarn)
	findings, err := checker.Scan(expirePath)
	if err != nil {
		return err
	}

	fmt.Println(expire.Format(findings))

	for _, f := range findings {
		if f.Expired {
			return fmt.Errorf("one or more secrets have expired")
		}
	}
	return nil
}
