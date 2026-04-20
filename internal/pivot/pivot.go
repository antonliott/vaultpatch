// Package pivot provides functionality to reorganise secrets by pivoting
// a key dimension — grouping all paths that share a given key name and
// collecting their values into a new synthetic secret.
package pivot

import (
	"context"
	"fmt"
	"sort"
)

// Reader lists and reads KV secrets.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Result holds the pivoted view: map[keyName]map[sourcePath]value.
type Result map[string]map[string]string

// Pivoter builds a Result from all secrets under a mount path.
type Pivоter struct {
	reader Reader
}

// NewPivoter returns a Pivоter backed by r.
func NewPivoter(r Reader) *Pivоter {
	return &Pivоter{reader: r}
}

// Apply reads every secret reachable from root and pivots by key name.
// The returned Result maps each discovered key to the set of paths that
// contain it, together with the value at that path.
func (p *Pivоter) Apply(ctx context.Context, root string) (Result, error) {
	paths, err := p.reader.List(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("pivot: list %q: %w", root, err)
	}

	out := make(Result)

	for _, path := range paths {
		secret, err := p.reader.Read(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("pivot: read %q: %w", path, err)
		}
		for k, v := range secret {
			if out[k] == nil {
				out[k] = make(map[string]string)
			}
			out[k][path] = v
		}
	}

	return out, nil
}

// Format renders a Result as a human-readable string.
func Format(r Result) string {
	if len(r) == 0 {
		return "(no keys found)\n"
	}

	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out string
	for _, k := range keys {
		out += fmt.Sprintf("[%s]\n", k)
		paths := make([]string, 0, len(r[k]))
		for path := range r[k] {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		for _, path := range paths {
			out += fmt.Sprintf("  %s = %s\n", path, r[k][path])
		}
	}
	return out
}
