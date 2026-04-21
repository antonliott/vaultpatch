package main

import (
	"fmt"
	"os"

	"github.com/example/vaultpatch/internal/linkage"
	"github.com/spf13/cobra"
)

func init() {
	var prefix string

	cmd := &cobra.Command{
		Use:   "linkage",
		Short: "Detect cross-path secret references within a Vault namespace",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLinkage(cmd, prefix)
		},
	}

	cmd.Flags().StringVar(&prefix, "prefix", "", "KV mount prefix to scan (required)")
	_ = cmd.MarkFlagRequired("prefix")

	rootCmd.AddCommand(cmd)
}

func runLinkage(cmd *cobra.Command, prefix string) error {
	client, err := newVaultClient()
	if err != nil {
		return fmt.Errorf("linkage: vault client: %w", err)
	}

	detector := linkage.NewDetector(client, prefix)
	links, err := detector.Detect(cmd.Context())
	if err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, linkage.Format(links))

	if len(links) > 0 {
		fmt.Fprintf(os.Stderr, "%d cross-path link(s) detected\n", len(links))
	}
	return nil
}
