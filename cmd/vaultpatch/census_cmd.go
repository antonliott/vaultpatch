package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/census"
	"github.com/example/vaultpatch/internal/vault"
)

var censusCmd = &cobra.Command{
	Use:   "census <path>",
	Short: "Report key distribution statistics across secret paths",
	Args:  cobra.ExactArgs(1),
	RunE:  runCensus,
}

func init() {
	rootCmd.AddCommand(censusCmd)
}

func runCensus(cmd *cobra.Command, args []string) error {
	root := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	col := census.NewCollector(client)
	report, err := col.Collect(context.Background(), root)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Paths: %d  Total keys: %d\n\n", report.TotalPaths, report.TotalKeys)

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tKEYS")
	for _, s := range report.Paths {
		fmt.Fprintf(w, "%s\t%d\n", s.Path, s.KeyCount)
	}
	w.Flush()

	if len(report.KeyFreq) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\nKey frequency:")
		wf := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(wf, "KEY\tCOUNT")
		for k, n := range report.KeyFreq {
			fmt.Fprintf(wf, "%s\t%d\n", k, n)
		}
		wf.Flush()
	}

	return nil
}
