package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/vaultpatch/internal/squash"
	"github.com/spf13/cobra"
)

func init() {
	var dryRun bool
	var dest string

	cmd := &cobra.Command{
		Use:   "squash --dest PATH SOURCE [SOURCE...]",
		Short: "Merge multiple secret paths into a single destination path",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSquash(dest, args, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	cmd.Flags().StringVar(&dest, "dest", "", "Destination secret path (required)")
	_ = cmd.MarkFlagRequired("dest")

	rootCmd.AddCommand(cmd)
}

func runSquash(dest string, sources []string, dryRun bool) error {
	client, err := newVaultClient()
	if err != nil {
		return err
	}

	s := squash.New(client, dryRun)
	r := s.Apply(dest, sources)

	if r.Err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", r.Err)
		return r.Err
	}

	if dryRun {
		fmt.Printf("[dry-run] would squash %s into %s\n",
			strings.Join(sources, ", "), dest)
	} else {
		fmt.Println(r.String())
	}
	return nil
}
