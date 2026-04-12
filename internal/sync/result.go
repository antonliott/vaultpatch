package sync

import "fmt"

// Result holds the outcome of a sync operation.
type Result struct {
	Synced []string
	Errors []error
	DryRun bool
}

// HasErrors returns true when at least one error was recorded.
func (r Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}

	s := fmt.Sprintf("[%s] synced %d secret(s)", mode, len(r.Synced))
	if r.HasErrors() {
		s += fmt.Sprintf(", %d error(s)", len(r.Errors))
	}
	return s
}
