package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/vaultpatch/internal/annotate"
	"github.com/spf13/cobra"
)

func init() {
	var (
		path       string
		addKVs     []string
		removeKeys []string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "annotate",
		Short: "Add, update, or remove annotations on a Vault secret path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnnotate(path, addKVs, removeKeys, dryRun)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "Vault secret path to annotate (required)")
	cmd.Flags().StringArrayVar(&addKVs, "set", nil, "Annotation to set in key=value format (repeatable)")
	cmd.Flags().StringArrayVar(&removeKeys, "remove", nil, "Annotation key to remove (repeatable)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}

func runAnnotate(path string, setKVs, removeKeys []string, dryRun bool) error {
	client, err := newVaultClient()
	if err != nil {
		return err
	}

	annotations := make(map[string]string, len(setKVs))
	for _, kv := range setKVs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set value %q: expected key=value", kv)
		}
		annotations[parts[0]] = parts[1]
	}

	a := annotate.NewAnnotator(client)
	res := a.Apply(rootCmd.Context(), annotate.Options{
		Path:        path,
		Annotations: annotations,
		RemoveKeys:  removeKeys,
		DryRun:      dryRun,
	})

	if res.Err != nil {
		return res.Err
	}

	if dryRun {
		fmt.Fprintf(os.Stdout, "[dry-run] annotate %s: +%d ~%d -%d\n", path, res.Added, res.Updated, res.Removed)
	} else {
		fmt.Fprintf(os.Stdout, "annotated %s: +%d ~%d -%d\n", path, res.Added, res.Updated, res.Removed)
	}
	return nil
}
