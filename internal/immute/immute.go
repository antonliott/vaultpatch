// Package immute provides functionality to mark secrets as immutable,
// preventing further writes to protected paths.
package immute

import (
	"context"
	"fmt"
)

// Writer writes key-value pairs to a secret path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]string) error
	Read(ctx context.Context, path string) (map[string]string, error)
}

// ImmutableKey is the metadata key used to mark a secret as immutable.
const ImmutableKey = "_immutable"

// Locker marks and checks immutability of secrets.
type Locker struct {
	rw  Writer
	dry bool
}

// NewLocker creates a new Locker.
func NewLocker(rw Writer, dryRun bool) *Locker {
	return &Locker{rw: rw, dry: dryRun}
}

// IsImmutable returns true if the secret at path is marked immutable.
func (l *Locker) IsImmutable(ctx context.Context, path string) (bool, error) {
	data, err := l.rw.Read(ctx, path)
	if err != nil {
		return false, fmt.Errorf("immute: read %q: %w", path, err)
	}
	return data[ImmutableKey] == "true", nil
}

// Apply marks the given paths as immutable by writing the sentinel key.
func (l *Locker) Apply(ctx context.Context, paths []string) (Result, error) {
	res := Result{DryRun: l.dry}
	for _, p := range paths {
		immutable, err := l.IsImmutable(ctx, p)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		if immutable {
			res.Skipped = append(res.Skipped, p)
			continue
		}
		if !l.dry {
			data, err := l.rw.Read(ctx, p)
			if err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", p, err))
				continue
			}
			data[ImmutableKey] = "true"
			if err := l.rw.Write(ctx, p, data); err != nil {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", p, err))
				continue
			}
		}
		res.Marked = append(res.Marked, p)
	}
	return res, nil
}
