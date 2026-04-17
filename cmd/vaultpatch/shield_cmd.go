package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/vaultpatch/internal/shield"
	"github.com/spf13/cobra"
)

func init() {
	var path string
	var keys []string
	var dryRun bool
	var data []string

	cmd := &cobra.Command{
		Use:   "shield",
		Short: "Write secrets while protecting specified keys from modification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShield(path, keys, data, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault secret path (required)")
	cmd.Flags().StringSliceVar(&keys, "protect", nil, "Keys to protect from overwrite")
	cmd.Flags().StringArrayVar(&data, "set", nil, "key=value pairs to write")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runShield(path string, protectKeys, pairs []string, dryRun bool) error {
	client, err := newVaultClient()
	if err != nil {
		return err
	}

	incoming := make(map[string]string, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set value %q: expected key=value", p)
		}
		incoming[parts[0]] = parts[1]
	}

	s := shield.NewShielder(client, protectKeys, dryRun)
	res, err := s.Apply(path, incoming)
	if err != nil {
		return err
	}

	if res.Blocked {
		fmt.Fprintf(os.Stderr, "shield: blocked keys (protected): %s\n", strings.Join(res.Protected, ", "))
	}
	if dryRun {
		fmt.Println("[dry-run] no changes written")
	} else {
		fmt.Printf("shield: wrote %s\n", path)
	}
	return nil
}
