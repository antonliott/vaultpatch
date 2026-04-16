package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/youorg/vaultpatch/internal/split"
	"github.com/youorg/vaultpatch/internal/vault"
)

var splitCmd = &cobra.Command{
	Use:   "split <src-path>",
	Short: "Split a secret into sub-paths keyed by a field value",
	Args:  cobra.ExactArgs(1),
	RunE:  runSplit,
}

var (
	splitKeyField string
	splitDestDir  string
	splitDryRun   bool
)

func init() {
	splitCmd.Flags().StringVar(&splitKeyField, "key-field", "env", "Secret field whose value determines the destination sub-path")
	splitCmd.Flags().StringVar(&splitDestDir, "dest", "", "Destination directory path (required)")
	splitCmd.Flags().BoolVar(&splitDryRun, "dry-run", false, "Preview changes without writing")
	_ = splitCmd.MarkFlagRequired("dest")
	rootCmd.AddCommand(splitCmd)
}

func runSplit(cmd *cobra.Command, args []string) error {
	src := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	s := split.NewSplitter(client, splitKeyField, splitDestDir, splitDryRun)
	res, err := s.Apply(src)
	if err != nil {
		return err
	}

	if splitDryRun {
		fmt.Println("[dry-run] would write:")
	} else {
		fmt.Println("written:")
	}
	for _, p := range res.Written {
		fmt.Println(" ", p)
	}
	for _, p := range res.Skipped {
		fmt.Println("  skipped (no key field):", strings.TrimSpace(p))
	}
	return nil
}
