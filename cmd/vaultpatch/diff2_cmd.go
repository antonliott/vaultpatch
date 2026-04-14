package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/example/vaultpatch/internal/diff2"
	"github.com/example/vaultpatch/internal/snapshot"
)

var diff2Cmd = &cobra.Command{
	Use:   "diff2 <snapshot-a> <snapshot-b>",
	Short: "Diff two snapshot files and print key-level changes",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff2,
}

func init() {
	rootCmd.AddCommand(diff2Cmd)
}

func runDiff2(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	_ = ctx

	snpA, err := snapshot.Load(args[0])
	if err != nil {
		return fmt.Errorf("loading snapshot A: %w", err)
	}
	snpB, err := snapshot.Load(args[1])
	if err != nil {
		return fmt.Errorf("loading snapshot B: %w", err)
	}

	src := flattenSnapshot(snpA)
	dst := flattenSnapshot(snpB)

	result := diff2.Compare(src, dst)
	fmt.Fprint(os.Stdout, diff2.Format(result))
	return nil
}

// flattenSnapshot converts a snapshot's Secrets map into the shape expected
// by diff2.Compare: map[path]map[key]value.
func flattenSnapshot(snp *snapshot.Snapshot) map[string]map[string]string {
	out := make(map[string]map[string]string, len(snp.Secrets))
	for path, kv := range snp.Secrets {
		copy := make(map[string]string, len(kv))
		for k, v := range kv {
			copy[k] = v
		}
		out[path] = copy
	}
	return out
}
