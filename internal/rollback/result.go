package rollback

import "fmt"

// PathError pairs a secret path with the error that occurred.
type PathError struct {
	Path string
	Err  error
}

// Result holds the outcome of a rollback operation.
type Result struct {
	DryRun      bool
	Reverted    []string
	WouldRevert []string
	Skipped     int
	Errors      []PathError
}

func (r *Result) addError(path string, err error) {
	r.Errors = append(r.Errors, PathError{Path: path, Err: err})
}

// HasErrors returns true if any errors were recorded.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a human-readable summary of the result.
func (r *Result) Summary() string {
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would revert %d path(s), %d already current, %d error(s)",
			len(r.WouldRevert), r.Skipped, len(r.Errors))
	}
	return fmt.Sprintf("reverted %d path(s), %d already current, %d error(s)",
		len(r.Reverted), r.Skipped, len(r.Errors))
}
