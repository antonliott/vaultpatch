package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/clone"
	"github.com/your-org/vaultpatch/internal/vault"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <src-path> <dst-path>",
	Short: "Clone secrets from one Vault path to another",
	Args:  cobra.ExactArgs(2),
	RunE:  runClone,
}

func init() {
	cloneCmd.Flags().Bool("dry-run", false, "Print what would be cloned without writing")
	cloneCmd.Flags().String("src-namespace", "", "Source Vault namespace (overrides VAULT_NAMESPACE)")
	cloneCmd.Flags().String("dst-namespace", "", "Destination Vault namespace")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	srcNS, _ := cmd.Flags().GetString("src-namespace")
	dstNS, _ := cmd.Flags().GetString("dst-namespace")

	if srcNS == "" {
		srcNS = os.Getenv("VAULT_NAMESPACE")
	}

	srcClient, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: srcNS,
	})
	if err != nil {
		return fmt.Errorf("source vault client: %w", err)
	}

	dstClient, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: dstNS,
	})
	if err != nil {
		return fmt.Errorf("destination vault client: %w", err)
	}

	c := clone.NewCloner(srcClient, dstClient, dryRun)
	res := c.Apply(context.Background(), srcPath, dstPath)

	log.Println(res.Summary())

	if res.Err != nil {
		return res.Err
	}
	return nil
}
