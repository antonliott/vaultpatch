package main

import (
	"fmt"
	"os"

	"github.com/example/vaultpatch/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchMount      string
	searchKeyPat     string
	searchValuePat   string
	searchMaskValues bool
)

func init() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search secrets by key or value pattern",
		RunE:  runSearch,
	}
	searchCmd.Flags().StringVar(&searchMount, "mount", "secret", "KV mount path")
	searchCmd.Flags().StringVar(&searchKeyPat, "key", "", "regex to match secret keys")
	searchCmd.Flags().StringVar(&searchValuePat, "value", "", "regex to match secret values")
	searchCmd.Flags().BoolVar(&searchMaskValues, "mask", true, "mask secret values in output")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, _ []string) error {
	client, err := buildClient()
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	s, err := search.NewSearcher(client, searchMount, searchKeyPat, searchValuePat)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	matches, err := s.Run(cmd.Context())
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	search.Format(os.Stdout, matches, searchMaskValues)
	return nil
}
