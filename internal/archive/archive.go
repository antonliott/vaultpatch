// Package archive provides functionality for archiving and retrieving
// soft-deleted secrets across Vault namespaces.
package archive

import (
	"context"
	"fmt"
	"time"
)

// Writer can write a secret to an archive path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]interface{}) error
}

// Reader can read a secret from a path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]interface{}, error)
	List(ctx context.Context, path string) ([]string, error)
}

// Archiver moves secrets to an archive prefix instead of deleting them.
type Archiver struct {
	reader      Reader
	writer      Writer
	archiveRoot string
	dryRun      bool
}

// NewArchiver creates an Archiver that stores archived secrets under archiveRoot.
func NewArchiver(r Reader, w Writer, archiveRoot string, dryRun bool) *Archiver {
	return &Archiver{reader: r, writer: w, archiveRoot: archiveRoot, dryRun: dryRun}
}

// Result holds the outcome of an archive operation.
type Result struct {
	Archived []string
	Skipped  []string
	Errors   []error
	DryRun   bool
}

// HasErrors returns true if any errors were recorded.
func (r *Result) HasErrors() bool { return len(r.Errors) > 0 }

// Apply reads secrets from sourcePaths and writes them under archiveRoot,
// tagging each entry with an archived_at timestamp.
func (a *Archiver) Apply(ctx context.Context, sourcePaths []string) (*Result, error) {
	res := &Result{DryRun: a.dryRun}
	ts := time.Now().UTC().Format(time.RFC3339)

	for _, src := range sourcePaths {
		data, err := a.reader.Read(ctx, src)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("read %s: %w", src, err))
			continue
		}
		if data == nil {
			res.Skipped = append(res.Skipped, src)
			continue
		}

		archived := make(map[string]interface{}, len(data)+1)
		for k, v := range data {
			archived[k] = v
		}
		archived["_archived_at"] = ts

		dest := fmt.Sprintf("%s/%s", a.archiveRoot, src)
		if !a.dryRun {
			if err := a.writer.Write(ctx, dest, archived); err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("write %s: %w", dest, err))
				continue
			}
		}
		res.Archived = append(res.Archived, src)
	}
	return res, nil
}
