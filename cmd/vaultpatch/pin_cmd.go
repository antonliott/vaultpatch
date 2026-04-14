package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/pin"
	"github.com/your-org/vaultpatch/internal/vault"
)

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin and drift-check secrets",
}

var pinCaptureCmd = &cobra.Command{
	Use:   "capture <path>",
	Short: "Capture current secret state as a pin file",
	Args:  cobra.ExactArgs(1),
	RunE:  runPinCapture,
}

var pinDriftCmd = &cobra.Command{
	Use:   "drift <pin-file>",
	Short: "Check live secret against a pin file",
	Args:  cobra.ExactArgs(1),
	RunE:  runPinDrift,
}

var pinRestoreCmd = &cobra.Command{
	Use:   "restore <pin-file>",
	Short: "Restore a secret to its pinned state",
	Args:  cobra.ExactArgs(1),
	RunE:  runPinRestore,
}

var pinOutput string
var pinDryRun bool

func init() {
	pinCaptureCmd.Flags().StringVarP(&pinOutput, "output", "o", "", "write pin JSON to file (default: stdout)")
	pinRestoreCmd.Flags().BoolVar(&pinDryRun, "dry-run", false, "preview restore without writing")
	pinCmd.AddCommand(pinCaptureCmd, pinDriftCmd, pinRestoreCmd)
	rootCmd.AddCommand(pinCmd)
}

func runPinCapture(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient(vault.ConfigFromEnv())
	if err != nil {
		return err
	}
	p := pin.NewPinner(client, client)
	entry, err := p.Pin(cmd.Context(), args[0])
	if err != nil {
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	if pinOutput != "" {
		return os.WriteFile(pinOutput, data, 0o600)
	}
	fmt.Println(string(data))
	return nil
}

func runPinDrift(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("pin drift: read file: %w", err)
	}
	var entry pin.PinEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return fmt.Errorf("pin drift: parse: %w", err)
	}
	client, err := vault.NewClient(vault.ConfigFromEnv())
	if err != nil {
		return err
	}
	p := pin.NewPinner(client, client)
	result, err := p.CheckDrift(cmd.Context(), &entry)
	if err != nil {
		return err
	}
	fmt.Print(pin.FormatDrift([]*pin.DriftResult{result}))
	if result.Drifted {
		os.Exit(1)
	}
	return nil
}

func runPinRestore(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("pin restore: read file: %w", err)
	}
	var entry pin.PinEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return fmt.Errorf("pin restore: parse: %w", err)
	}
	client, err := vault.NewClient(vault.ConfigFromEnv())
	if err != nil {
		return err
	}
	p := pin.NewPinner(client, client)
	if err := p.Restore(cmd.Context(), &entry, pinDryRun); err != nil {
		return err
	}
	if pinDryRun {
		fmt.Println("dry-run: restore skipped")
	} else {
		fmt.Printf("restored %s to pinned state\n", entry.Path)
	}
	return nil
}
