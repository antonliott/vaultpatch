package immute

import "fmt"

// Result holds the outcome of an immutability Apply operation.
type Result struct {
	DryRun  bool
	Marked  []string
	Skipped []string
	Errors  []string
}

// HasErrors returns true if any errors occurred.
func (r Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	prefix := ""
	if r.DryRun {
		prefix = "[dry-run] "
	}
	return fmt.Sprintf("%smarked=%d skipped=%d errors=%d",
		prefix, len(r.Marked), len(r.Skipped), len(r.Errors))
}
