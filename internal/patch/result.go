package patch

import (
	"fmt"
	"strings"
	"time"
)

// Result holds the outcome of a patch operation.
type Result struct {
	Path      string
	Applied   int
	Skipped   int
	DryRun    bool
	AppliedAt time.Time
	Errors    []error
}

// HasErrors returns true if any errors were recorded.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddError appends an error to the result.
func (r *Result) AddError(err error) {
	r.Errors = append(r.Errors, err)
}

// Summary returns a human-readable summary of the result.
func (r *Result) Summary() string {
	var sb strings.Builder
	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(&sb, "[%s] path=%s applied=%d skipped=%d",
		mode, r.Path, r.Applied, r.Skipped)
	if r.HasErrors() {
		fmt.Fprintf(&sb, " errors=%d", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Fprintf(&sb, "\n  - %s", e.Error())
		}
	}
	return sb.String()
}
