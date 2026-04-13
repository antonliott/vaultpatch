package search

import (
	"fmt"
	"io"
	"strings"
)

// Format writes matches to w in a human-readable table.
// If maskValues is true, secret values are replaced with "***".
func Format(w io.Writer, matches []Match, maskValues bool) {
	if len(matches) == 0 {
		fmt.Fprintln(w, "no matches found")
		return
	}
	fmt.Fprintf(w, "%-40s %-24s %s\n", "PATH", "KEY", "VALUE")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	for _, m := range matches {
		v := m.Value
		if maskValues {
			v = "***"
		}
		fmt.Fprintf(w, "%-40s %-24s %s\n", m.Path, m.Key, v)
	}
}
