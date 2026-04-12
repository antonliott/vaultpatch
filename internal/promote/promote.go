// Package promote handles copying secrets from one Vault namespace to another.
package promote

import (
	"context"
	"fmt"
)

// SecretReader reads a secret at a given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// SecretWriter writes a secret at a given path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of a promotion operation.
type Result struct {
	Path    string
	Skipped bool
	Err     error
}

// Promoter copies secrets from a source to a destination.
type Promoter struct {
	src    SecretReader
	dst    SecretWriter
	dryRun bool
}

// NewPromoter creates a Promoter.
func NewPromoter(src SecretReader, dst SecretWriter, dryRun bool) *Promoter {
	return &Promoter{src: src, dst: dst, dryRun: dryRun}
}

// Apply promotes each path from source to destination.
func (p *Promoter) Apply(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, p.promote(ctx, path))
	}
	return results
}

func (p *Promoter) promote(ctx context.Context, path string) Result {
	data, err := p.src.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read %q: %w", path, err)}
	}
	if p.dryRun {
		return Result{Path: path, Skipped: true}
	}
	if err := p.dst.WriteSecret(ctx, path, data); err != nil {
		return Result{Path: path, Err: fmt.Errorf("write %q: %w", path, err)}
	}
	return Result{Path: path}
}

// Summary returns a human-readable summary of promotion results.
func Summary(results []Result, dryRun bool) string {
	var ok, skipped, failed int
	for _, r := range results {
		switch {
		case r.Err != nil:
			failed++
		case r.Skipped:
			skipped++
		default:
			ok++
		}
	}
	if dryRun {
		return fmt.Sprintf("dry-run: %d path(s) would be promoted, %d error(s)", skipped, failed)
	}
	return fmt.Sprintf("promoted %d path(s), %d skipped, %d error(s)", ok, skipped, failed)
}
