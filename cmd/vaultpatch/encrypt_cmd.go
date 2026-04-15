package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vaultpatch/vaultpatch/internal/encrypt"
	"github.com/vaultpatch/vaultpatch/internal/vault"
)

func init() {
	var (
		path    string
		keyHex  string
		decrypt bool
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt or decrypt secret values at a Vault path using AES-GCM",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEncrypt(path, keyHex, decrypt, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault KV path (required)")
	cmd.Flags().StringVar(&keyHex, "key", "", "Hex-encoded AES key (16, 24, or 32 bytes) (required)")
	cmd.Flags().BoolVar(&decrypt, "decrypt", false, "Decrypt values instead of encrypting")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print result without writing back to Vault")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("key")

	rootCmd.AddCommand(cmd)
}

func runEncrypt(path, keyHex string, decrypt, dryRun bool) error {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid key hex: %w", err)
	}

	enc, err := encrypt.New(keyBytes)
	if err != nil {
		return err
	}

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return err
	}

	secrets, err := client.Read(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	var result map[string]string
	if decrypt {
		result, err = enc.Decrypt(secrets)
	} else {
		result, err = enc.Encrypt(secrets)
	}
	if err != nil {
		return err
	}

	if dryRun {
		action := "encrypt"
		if decrypt {
			action = "decrypt"
		}
		fmt.Fprintf(os.Stdout, "[dry-run] would %s %d keys at %q\n", action, len(result), path)
		for k := range result {
			fmt.Fprintf(os.Stdout, "  %s\n", k)
		}
		return nil
	}

	if err := client.Write(path, result); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	fmt.Fprintf(os.Stdout, "updated %d keys at %q\n", len(result), path)
	return nil
}
