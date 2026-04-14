// Package scope provides utilities for listing and filtering Vault secret
// paths within a namespace, supporting glob-style prefix matching.
package scope

import (
	"context"
	"fmt"
	"path"
	"strings"
)

// Reader lists secret keys under a given mount and path prefix.
type Reader interface {
	List(ctx context.Context, mountPath, prefix string) ([]string, error)
}

// Filter holds options for narrowing which paths are included.
type Filter struct {
	// Prefix restricts results to paths beginning with this value.
	Prefix string
	// Exclude is a list of exact path segments to skip.
	Exclude []string
}

// Scope resolves a set of secret paths within a Vault mount.
type Scope struct {
	reader Reader
	mount  string
}

// NewScope creates a Scope backed by the given Reader for the specified mount.
func NewScope(r Reader, mount string) *Scope {
	return &Scope{reader: r, mount: mount}
}

// Resolve returns all secret paths under the mount that satisfy f.
func (s *Scope) Resolve(ctx context.Context, f Filter) ([]string, error) {
	all, err := s.reader.List(ctx, s.mount, f.Prefix)
	if err != nil {
		return nil, fmt.Errorf("scope: list %q/%q: %w", s.mount, f.Prefix, err)
	}

	excluded := make(map[string]struct{}, len(f.Exclude))
	for _, e := range f.Exclude {
		excluded[strings.Trim(e, "/")] = struct{}{}
	}

	var results []string
	for _, p := range all {
		clean := path.Clean(p)
		if _, skip := excluded[clean]; skip {
			continue
		}
		results = append(results, clean)
	}
	return results, nil
}
