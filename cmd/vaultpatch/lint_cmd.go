package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/youorg/vaultpatch/internal/lint"
	"github.com/youorg/vaultpatch/internal/vault"
)

func init() {
	var path string
	var failOnWarn bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint secrets at a Vault path against built-in rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLint(path, failOnWarn)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault KV path to lint (required)")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit non-zero on warnings as well as errors")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runLint(path string, failOnWarn bool) error {
	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	secrets, err := client.Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	l := lint.New(nil)
	findings := l.Run(path, secrets)

	if len(findings) == 0 {
		fmt.Println("✔ no lint findings")
		return nil
	}

	for _, f := range findings {
		icon := "⚠"
		if f.Severity == lint.SeverityError {
			icon = "✖"
		}
		fmt.Fprintf(os.Stderr, "%s [%s] %s › %s: %s\n", icon, f.Severity, f.Path, f.Key, f.Message)
	}

	if lint.HasErrors(findings) {
		return fmt.Errorf("lint failed with %d error(s)", countBySeverity(findings, lint.SeverityError))
	}
	if failOnWarn && len(findings) > 0 {
		return fmt.Errorf("lint failed with %d warning(s)", len(findings))
	}
	return nil
}

func countBySeverity(findings []lint.Finding, s lint.Severity) int {
	n := 0
	for _, f := range findings {
		if f.Severity == s {
			n++
		}
	}
	return n
}
