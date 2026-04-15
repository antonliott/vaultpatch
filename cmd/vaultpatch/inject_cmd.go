package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/inject"
)

func init() {
	cmd := &cobra.Command{
		Use:   "inject",
		Short: "Resolve vault:// references in a JSON key-value file",
		RunE:  runInject,
	}
	cmd.Flags().StringP("file", "f", "", "path to JSON key-value file (required)")
	cmd.Flags().Bool("dry-run", false, "print resolved map without writing")
	_ = cmd.MarkFlagRequired("file")
	rootCmd.AddCommand(cmd)
}

func runInject(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("inject: read file: %w", err)
	}

	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("inject: parse JSON: %w", err)
	}

	vc, err := newVaultClient(cmd)
	if err != nil {
		return err
	}

	inj := inject.New(vc, inject.Options{DryRun: dryRun})
	out, res := inj.Apply(context.Background(), data)

	if res.HasErrors() {
		for _, e := range res.Errors {
			fmt.Fprintf(os.Stderr, "warn: %v\n", e)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("inject: encode output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "injected=%d skipped=%d errors=%d\n",
		res.Injected, res.Skipped, len(res.Errors))
	return nil
}
