package expire

import (
	"fmt"
	"strings"
	"time"
)

// Format returns a human-readable summary of findings.
func Format(findings []Finding) string {
	if len(findings) == 0 {
		return "no expiring secrets found"
	}
	var sb strings.Builder
	now := time.Now().UTC()
	for _, f := range findings {
		status := "EXPIRING"
		if f.Expired {
			status = "EXPIRED"
		}
		delta := f.Expires.Sub(now).Truncate(time.Second)
		if f.Expired {
			delta = now.Sub(f.Expires).Truncate(time.Second)
			fmt.Fprintf(&sb, "[%s] %s — expired %s ago (%s)\n", status, f.Path, delta, f.Expires.Format(time.RFC3339))
		} else {
			fmt.Fprintf(&sb, "[%s] %s — expires in %s (%s)\n", status, f.Path, delta, f.Expires.Format(time.RFC3339))
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}
