package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/reorder"
	"github.com/example/vaultpatch/internal/vault"
)

func init() {
	var (
		path    string
		order   string
		reverse bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder keys within a Vault secret",
		Long: `Reorder keys within a Vault KV secret.

Keys can be sorted alphabetically (default), in reverse order (--reverse),
or according to an explicit comma-separated list (--order).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReorder(path, order, reverse, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault secret path (required)")
	cmd.Flags().StringVar(&order, "order", "", "Comma-separated explicit key order")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Sort keys in reverse alphabetical order")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runReorder(path, orderFlag string, reverse, dryRun bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	var explicitOrder []string
	if orderFlag != "" {
		for _, k := range strings.Split(orderFlag, ",") {
			if t := strings.TrimSpace(k); t != "" {
				explicitOrder = append(explicitOrder, t)
			}
		}
	}

	r := reorder.New(client, reorder.Options{
		Order:   explicitOrder,
		Reverse: reverse,
		DryRun:  dryRun,
	})

	res := r.Apply(cmd.Context(), path)
	fmt.Println(reorder.Format(res))
	if res.Err != nil {
		return res.Err
	}
	return nil
}
