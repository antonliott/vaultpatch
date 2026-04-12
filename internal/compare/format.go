package compare

import (
	"fmt"
	"strings"
)

// Format renders a Result as a human-readable diff string.
func Format(r Result) string {
	if !r.HasDiff() {
		return "(no differences)"
	}

	var sb strings.Builder
	for _, d := range r.Deltas {
		switch d.Kind {
		case Added:
			fmt.Fprintf(&sb, "+ %s#%s = %v\n", d.Path, d.Key, d.SourceVal)
		case Removed:
			fmt.Fprintf(&sb, "- %s#%s (was %v)\n", d.Path, d.Key, d.TargetVal)
		case Changed:
			fmt.Fprintf(&sb, "~ %s#%s: %v -> %v\n", d.Path, d.Key, d.TargetVal, d.SourceVal)
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}
