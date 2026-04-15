package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/inherit"
	"github.com/your-org/vaultpatch/internal/vault"
)

var inheritCmd = &cobra.Command{
	Use:   "inherit <parent-path> <child-path>[,<child-path>...]",
	Short: "Propagate secrets from a parent path to child paths",
	Args:  cobra.ExactArgs(2),
	RunE:  runInherit,
}

func init() {
	inheritCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	inheritCmd.Flags().Bool("overwrite", false, "Overwrite existing keys in child paths")
	rootCmd.AddCommand(inheritCmd)
}

func runInherit(cmd *cobra.Command, args []string) error {
	parent := args[0]
	children := strings.Split(args[1], ",")

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	overwrite, _ := cmd.Flags().GetBool("overwrite")

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	inheritor := inherit.NewInheritor(client, inherit.Options{
		DryRun:    dryRun,
		Overwrite: overwrite,
	})

	result, err := inheritor.Apply(context.Background(), parent, children)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "[dry-run] no changes written")
	}

	for _, cr := range result.Children {
		if cr.Err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error %s: %v\n", cr.Path, cr.Err)
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s: applied=%d skipped=%d\n", cr.Path, cr.Applied, cr.Skipped)
	}

	return nil
}
