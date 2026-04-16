// Package resolve provides reference resolution across Vault paths,
// expanding placeholder values of the form {{path#key}} with live secret data.
package resolve

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Reader reads secret key/value pairs from a Vault path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Resolver expands {{path#key}} placeholders in secret values.
type Resolver struct {
	reader Reader
	pattern *regexp.Regexp
}

// New returns a Resolver backed by the given Reader.
func New(r Reader) *Resolver {
	return &Resolver{
		reader:  r,
		pattern: regexp.MustCompile(`\{\{([^#}]+)#([^}]+)\}\}`),
	}
}

// Apply resolves all placeholders in values of secrets at path.
// Returns a new map with references expanded; the original is not mutated.
func (r *Resolver) Apply(ctx context.Context, path string, dryRun bool) (map[string]string, int, error) {
	secrets, err := r.reader.Read(ctx, path)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve: read %q: %w", path, err)
	}

	out := make(map[string]string, len(secrets))
	cache := make(map[string]map[string]string)
	resolved := 0

	for k, v := range secrets {
		expanded, n, err := r.expandValue(ctx, v, cache)
		if err != nil {
			return nil, 0, fmt.Errorf("resolve: key %q: %w", k, err)
		}
		out[k] = expanded
		resolved += n
	}

	return out, resolved, nil
}

func (r *Resolver) expandValue(ctx context.Context, val string, cache map[string]map[string]string) (string, int, error) {
	var expandErr error
	count := 0

	result := r.pattern.ReplaceAllStringFunc(val, func(match string) string {
		if expandErr != nil {
			return match
		}
		subs := r.pattern.FindStringSubmatch(match)
		refPath, refKey := strings.TrimSpace(subs[1]), strings.TrimSpace(subs[2])

		data, ok := cache[refPath]
		if !ok {
			var err error
			data, err = r.reader.Read(ctx, refPath)
			if err != nil {
				expandErr = fmt.Errorf("read ref path %q: %w", refPath, err)
				return match
			}
			cache[refPath] = data
		}

		v, found := data[refKey]
		if !found {
			expandErr = fmt.Errorf("key %q not found in %q", refKey, refPath)
			return match
		}
		count++
		return v
	})

	if expandErr != nil {
		return "", 0, expandErr
	}
	return result, count, nil
}
