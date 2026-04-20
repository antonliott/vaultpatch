package reorder

import (
	"fmt"
	"strings"
)

// Format returns a human-readable summary of the reorder result.
func Format(r Result) string {
	if r.Err != nil {
		return fmt.Sprintf("error: %v", r.Err)
	}
	if len(r.Deltas) == 0 {
		return "no key order changes"
	}
	var sb strings.Builder
	if r.DryRun {
		sb.WriteString("[dry-run] ")
	}
	fmt.Fprintf(&sb, "%d key(s) reordered:\n", len(r.Deltas))
	for _, d := range r.Deltas {
		fmt.Fprintf(&sb, "  %s  %s  %d -> %d\n", d.Path, d.Key, d.From, d.To)
	}
	return strings.TrimRight(sb.String(), "\n")
}
