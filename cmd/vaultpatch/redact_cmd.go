package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/redact"
)

func init() {
	var keys []string
	var useDefaults bool

	cmd := &cobra.Command{
		Use:   "redact",
		Short: "Print secrets with sensitive values replaced by [REDACTED]",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRedact(keys, useDefaults)
		},
	}

	cmd.Flags().StringSliceVar(&keys, "keys", nil, "comma-separated list of keys to redact")
	cmd.Flags().BoolVar(&useDefaults, "defaults", true, "include built-in sensitive key patterns")

	rootCmd.AddCommand(cmd)
}

func runRedact(extraKeys []string, useDefaults bool) error {
	var r *redact.Redactor
	if useDefaults {
		r = redact.NewDefault()
	} else {
		r = redact.New(extraKeys)
	}

	// Read JSON map from stdin
	var secrets map[string]string
	if err := json.NewDecoder(os.Stdin).Decode(&secrets); err != nil {
		return fmt.Errorf("reading secrets from stdin: %w", err)
	}

	safe := r.Apply(secrets)

	// Print as KEY=VALUE lines
	for k, v := range safe {
		fmt.Printf("%s=%s\n", k, strings.ReplaceAll(v, "\n", "\\n"))
	}

	fmt.Fprintln(os.Stderr, r.Summary(secrets))
	return nil
}
