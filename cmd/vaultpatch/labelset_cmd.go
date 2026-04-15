package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/labelset"
)

var labelsetCmd = &cobra.Command{
	Use:   "labelset",
	Short: "Compare label sets between two secret paths",
}

func init() {
	var srcPath, dstPath string
	var srcLabels, dstLabels []string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff label sets between two paths (supplied inline)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLabelsetDiff(srcPath, dstPath, srcLabels, dstLabels)
		},
	}

	cmd.Flags().StringVar(&srcPath, "src", "", "Source secret path")
	cmd.Flags().StringVar(&dstPath, "dst", "", "Destination secret path")
	cmd.Flags().StringArrayVar(&srcLabels, "src-label", nil, "Source labels as key=value (repeatable)")
	cmd.Flags().StringArrayVar(&dstLabels, "dst-label", nil, "Destination labels as key=value (repeatable)")
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	labelsetCmd.AddCommand(cmd)
	rootCmd.AddCommand(labelsetCmd)
}

func parseInlineLabels(pairs []string) (labelset.Labels, error) {
	labels := make(labelset.Labels, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label %q: expected key=value", p)
		}
		labels[parts[0]] = parts[1]
	}
	return labels, nil
}

func runLabelsetDiff(srcPath, dstPath string, srcPairs, dstPairs []string) error {
	src, err := parseInlineLabels(srcPairs)
	if err != nil {
		return fmt.Errorf("src labels: %w", err)
	}
	dst, err := parseInlineLabels(dstPairs)
	if err != nil {
		return fmt.Errorf("dst labels: %w", err)
	}

	deltas := labelset.Compare(src, dst)
	fmt.Fprintf(os.Stdout, "Label diff: %s -> %s\n", srcPath, dstPath)
	fmt.Fprint(os.Stdout, labelset.Format(deltas))
	return nil
}
