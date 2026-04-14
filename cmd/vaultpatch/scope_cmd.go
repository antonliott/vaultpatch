package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/scope"
)

func init() {
	var (
		mount   string
		prefix  string
		exclude []string
	)

	cmd := &cobra.Command{
		Use:   "scope",
		Short: "List secret paths within a Vault mount, with optional filtering",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScope(mount, prefix, exclude)
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount to scope")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Path prefix to filter results")
	cmd.Flags().StringArrayVar(&exclude, "exclude", nil, "Paths to exclude (repeatable)")

	rootCmd.AddCommand(cmd)
}

func runScope(mount, prefix string, exclude []string) error {
	client, err := newVaultClient()
	if err != nil {
		return fmt.Errorf("scope: %w", err)
	}

	s := scope.NewScope(client, mount)
	paths, err := s.Resolve(rootCmd.Context(), scope.Filter{
		Prefix:  prefix,
		Exclude: exclude,
	})
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "no paths found")
		return nil
	}

	fmt.Println(strings.Join(paths, "\n"))
	return nil
}
