package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/tag"
)

var tagCmd = &cobra.Command{
	Use:   "tag <path>",
	Short: "Apply metadata tags to a Vault secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runTag,
}

func init() {
	tagCmd.Flags().StringToStringP("set", "s", nil, "Tags to set in key=value format (repeatable)")
	tagCmd.Flags().BoolP("dry-run", "n", false, "Preview changes without writing to Vault")
	rootCmd.AddCommand(tagCmd)
}

func runTag(cmd *cobra.Command, args []string) error {
	path := args[0]

	desiredRaw, err := cmd.Flags().GetStringToString("set")
	if err != nil {
		return fmt.Errorf("reading --set flags: %w", err)
	}
	if len(desiredRaw) == 0 {
		return fmt.Errorf("at least one --set key=value is required")
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	client, err := newVaultClient()
	if err != nil {
		return err
	}

	current, err := client.ReadMetadata(path)
	if err != nil {
		return fmt.Errorf("reading metadata for %s: %w", path, err)
	}

	desired := tag.Tags(desiredRaw)
	applier := tag.NewApplier(client, dryRun)
	result := applier.Apply(path, tag.Tags(current), desired)

	if result.Err != nil {
		return result.Err
	}

	out := map[string]interface{}{
		"path":    result.Path,
		"dry_run": result.DryRun,
		"applied": result.Applied,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(out); encErr != nil {
		return fmt.Errorf("encoding output: %w", encErr)
	}
	return nil
}
