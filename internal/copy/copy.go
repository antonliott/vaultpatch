// Package copy provides functionality to copy secrets between paths within
// or across Vault namespaces.
package copy

import (
	"context"
	"fmt"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secrets to a Vault path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Result holds the outcome of a copy operation.
type Result struct {
	Source      string
	Destination string
	KeysCopied  int
	DryRun      bool
	Err         error
}

// Copier copies secrets from one path to another.
type Copier struct {
	src    Reader
	dst    Writer
	dryRun bool
}

// NewCopier creates a new Copier.
func NewCopier(src Reader, dst Writer, dryRun bool) *Copier {
	return &Copier{src: src, dst: dst, dryRun: dryRun}
}

// Apply reads secrets from src and writes them to dst.
// If dryRun is true, no write is performed.
func (c *Copier) Apply(ctx context.Context, srcPath, dstPath string) Result {
	result := Result{
		Source:      srcPath,
		Destination: dstPath,
		DryRun:      c.dryRun,
	}

	data, err := c.src.Read(ctx, srcPath)
	if err != nil {
		result.Err = fmt.Errorf("read %q: %w", srcPath, err)
		return result
	}

	result.KeysCopied = len(data)

	if c.dryRun {
		return result
	}

	if err := c.dst.Write(ctx, dstPath, data); err != nil {
		result.Err = fmt.Errorf("write %q: %w", dstPath, err)
	}

	return result
}

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	if r.Err != nil {
		return fmt.Sprintf("error copying %q -> %q: %v", r.Source, r.Destination, r.Err)
	}
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would copy %d key(s) from %q to %q", r.KeysCopied, r.Source, r.Destination)
	}
	return fmt.Sprintf("copied %d key(s) from %q to %q", r.KeysCopied, r.Source, r.Destination)
}
