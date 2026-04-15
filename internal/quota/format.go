package quota

import (
	"fmt"
	"strings"
)

// Format renders quota violations as a human-readable report.
// Returns an empty string when there are no violations.
func Format(violations []Violation) string {
	if len(violations) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("quota violations (%d):\n", len(violations)))
	for _, v := range violations {
		sb.WriteString(fmt.Sprintf("  ! %s\n", v.String()))
	}
	return strings.TrimRight(sb.String(), "\n")
}
