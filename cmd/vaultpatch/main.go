// Package main is the entry point for the vaultpatch CLI tool.
// vaultpatch diffs and applies secrets changes across HashiCorp Vault namespaces.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// version is set at build time via ldflags
	version = "dev"

	// Global flags
	vaultAddr      string
	vaultToken     string
	vaultNamespace string
	outputFormat   string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// rootCmd is the base command for the vaultpatch CLI.
var rootCmd = &cobra.Command{
	Use:   "vaultpatch",
	Short: "Diff and apply secrets changes across HashiCorp Vault namespaces",
	Long: `vaultpatch is a CLI tool for comparing and synchronizing secrets
across HashiCorp Vault namespaces. It supports diffing secret paths,
reviewing changes, and applying patches safely.`,
	Version: version,
	SilenceUsage: true,
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(
		&vaultAddr,
		"vault-addr",
		getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		"Vault server address (env: VAULT_ADDR)",
	)

	rootCmd.PersistentFlags().StringVar(
		&vaultToken,
		"vault-token",
		getEnvOrDefault("VAULT_TOKEN", ""),
		"Vault authentication token (env: VAULT_TOKEN)",
	)

	rootCmd.PersistentFlags().StringVar(
		&vaultNamespace,
		"namespace",
		getEnvOrDefault("VAULT_NAMESPACE", ""),
		"Vault namespace (env: VAULT_NAMESPACE)",
	)

	rootCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		"text",
		"Output format: text, json, yaml",
	)

	// Register subcommands
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(applyCmd)
}

// diffCmd placeholder — implemented in diff.go
var diffCmd = &cobra.Command{
	Use:   "diff <source-path> <target-path>",
	Short: "Show differences between two Vault secret paths",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement in internal/diff package
		fmt.Fprintf(cmd.OutOrStdout(), "diff %s -> %s (not yet implemented)\n", args[0], args[1])
		return nil
	},
}

// applyCmd placeholder — implemented in apply.go
var applyCmd = &cobra.Command{
	Use:   "apply <patch-file>",
	Short: "Apply a secrets patch file to Vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement in internal/patch package
		fmt.Fprintf(cmd.OutOrStdout(), "apply %s (not yet implemented)\n", args[0])
		return nil
	},
}

// getEnvOrDefault returns the value of an environment variable,
// or the provided default if the variable is not set.
func getEnvOrDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
