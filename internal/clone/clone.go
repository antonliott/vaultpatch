// Package clone provides functionality to deep-copy secrets from one
// Vault path to another, optionally across namespaces.
package clone

import (
	"context"
	"fmt"
)

// SecretReader reads secret data from a given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// SecretWriter writes secret data to a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines read and write capabilities.
type ReadWriter interface {
	SecretReader
	SecretWriter
}

// Cloner copies secrets from a source path to a destination path.
type Cloner struct {
	src    SecretReader
	dst    SecretWriter
	dryRun bool
}

// NewCloner creates a new Cloner.
func NewCloner(src SecretReader, dst SecretWriter, dryRun bool) *Cloner {
	return &Cloner{src: src, dst: dst, dryRun: dryRun}
}

// Result holds the outcome of a clone operation.
type Result struct {
	SourcePath string
	DestPath   string
	Keys       []string
	DryRun     bool
	Err        error
}

// Apply reads secrets from srcPath and writes them to dstPath.
func (c *Cloner) Apply(ctx context.Context, srcPath, dstPath string) Result {
	res := Result{
		SourcePath: srcPath,
		DestPath:   dstPath,
		DryRun:     c.dryRun,
	}

	data, err := c.src.ReadSecret(ctx, srcPath)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", srcPath, err)
		return res
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	res.Keys = keys

	if c.dryRun {
		return res
	}

	if err := c.dst.WriteSecret(ctx, dstPath, data); err != nil {
		res.Err = fmt.Errorf("write %s: %w", dstPath, err)
	}
	return res
}

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	if r.Err != nil {
		return fmt.Sprintf("clone failed: %v", r.Err)
	}
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would clone %d key(s) from %s to %s", len(r.Keys), r.SourcePath, r.DestPath)
	}
	return fmt.Sprintf("cloned %d key(s) from %s to %s", len(r.Keys), r.SourcePath, r.DestPath)
}
