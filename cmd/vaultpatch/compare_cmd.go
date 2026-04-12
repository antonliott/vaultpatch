package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/vaultpatch/internal/compare"
	"github.com/example/vaultpatch/internal/vault"
	"github.com/spf13/cobra"
)

var (
	compareSrcNamespace string
	compareTgtNamespace string
	comparePath         string
)

func init() {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Diff secrets between two Vault namespaces",
		RunE:  runCompare,
	}
	cmd.Flags().StringVar(&compareSrcNamespace, "src-namespace", "", "Source Vault namespace (required)")
	cmd.Flags().StringVar(&compareTgtNamespace, "tgt-namespace", "", "Target Vault namespace (required)")
	cmd.Flags().StringVar(&comparePath, "path", "secret", "KV mount path to compare")
	_ = cmd.MarkFlagRequired("src-namespace")
	_ = cmd.MarkFlagRequired("tgt-namespace")
	rootCmd.AddCommand(cmd)
}

func runCompare(cmd *cobra.Command, _ []string) error {
	addr := getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200")
	token := os.Getenv("VAULT_TOKEN")

	srcClient, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: compareSrcNamespace,
	})
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}

	tgtClient, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: compareTgtNamespace,
	})
	if err != nil {
		return fmt.Errorf("target client: %w", err)
	}

	comp := compare.NewComparator(srcClient, tgtClient)
	result, err := comp.Compare(context.Background(), comparePath)
	if err != nil {
		return fmt.Errorf("compare: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), compare.Format(result))
	if result.HasDiff() {
		os.Exit(1)
	}
	return nil
}
