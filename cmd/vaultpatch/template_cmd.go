package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/your-org/vaultpatch/internal/template"
	"github.com/your-org/vaultpatch/internal/vault"
)

func init() {
	tmplCmd := &cobra.Command{
		Use:   "template",
		Short: "Render a template file against Vault secrets",
		RunE:  runTemplate,
	}
	tmplCmd.Flags().String("file", "", "Path to template file (required)")
	tmplCmd.Flags().String("path", "", "Vault KV secret path to read data from (required)")
	tmplCmd.Flags().String("mount", "secret", "KV v2 mount point")
	tmplCmd.Flags().String("out", "", "Output file path (default: stdout)")
	_ = tmplCmd.MarkFlagRequired("file")
	_ = tmplCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(tmplCmd)
}

func runTemplate(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	vaultPath, _ := cmd.Flags().GetString("path")
	mount, _ := cmd.Flags().GetString("mount")
	outPath, _ := cmd.Flags().GetString("out")

	tmplBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}

	client, err := vault.NewClient(vault.Config{
		Address:   getEnvOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	data, err := client.ReadSecret(cmd.Context(), mount, vaultPath)
	if err != nil {
		return fmt.Errorf("reading secret %q: %w", vaultPath, err)
	}

	strData := make(map[string]string, len(data))
	for k, v := range data {
		strData[k] = fmt.Sprintf("%v", v)
	}

	r := template.NewRenderer()
	result, err := r.Render(filePath, string(tmplBytes), strData)
	if err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	if outPath == "" {
		_, err = fmt.Fprint(cmd.OutOrStdout(), result)
		return err
	}

	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return os.WriteFile(outPath, []byte(result), 0o600)
}
