package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/rotate"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate secrets at specified Vault paths",
	RunE:  runRotate,
}

func init() {
	rotateCmd.Flags().StringP("file", "f", "", "JSON file containing rotation requests (required)")
	rotateCmd.Flags().Bool("dry-run", false, "Preview rotation without writing")
	_ = rotateCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open rotate file: %w", err)
	}
	defer f.Close()

	var reqs []rotate.RotateRequest
	if err := json.NewDecoder(f).Decode(&reqs); err != nil {
		return fmt.Errorf("decode rotate file: %w", err)
	}

	client, err := newVaultClient()
	if err != nil {
		return err
	}

	r := rotate.NewRotator(client, dryRun)
	results := r.Apply(context.Background(), reqs)

	hasErr := false
	for _, res := range results {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: %v\n", res.Path, res.Err)
			hasErr = true
		} else if res.DryRun {
			fmt.Printf("[dry-run] would rotate: %s\n", res.Path)
		} else {
			fmt.Printf("rotated: %s\n", res.Path)
		}
	}
	if hasErr {
		return fmt.Errorf("one or more rotation errors occurred")
	}
	return nil
}
