// Package inject provides functionality to inject secrets from Vault
// into a map of environment-variable-style key-value pairs, resolving
// Vault paths referenced as values.
package inject

import (
	"context"
	"fmt"
	"strings"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Options controls injection behaviour.
type Options struct {
	// Prefix is the marker prefix used to identify Vault references.
	// Defaults to "vault://".
	Prefix string
	DryRun bool
}

// Result holds the outcome of an injection run.
type Result struct {
	Injected int
	Skipped  int
	Errors   []error
}

func (r *Result) HasErrors() bool { return len(r.Errors) > 0 }

// Injector replaces vault:// references in a map with real secret values.
type Injector struct {
	reader Reader
	opts   Options
}

// New creates a new Injector.
func New(r Reader, opts Options) *Injector {
	if opts.Prefix == "" {
		opts.Prefix = "vault://"
	}
	return &Injector{reader: r, opts: opts}
}

// Apply scans values in data for vault:// references and replaces them.
// References have the form: vault://path/to/secret#key
func (inj *Injector) Apply(ctx context.Context, data map[string]string) (map[string]string, Result) {
	out := make(map[string]string, len(data))
	var res Result
	cache := map[string]map[string]string{}

	for k, v := range data {
		if !strings.HasPrefix(v, inj.opts.Prefix) {
			out[k] = v
			res.Skipped++
			continue
		}

		ref := strings.TrimPrefix(v, inj.opts.Prefix)
		path, field, ok := strings.Cut(ref, "#")
		if !ok || field == "" {
			res.Errors = append(res.Errors, fmt.Errorf("inject: invalid reference %q (missing #field)", v))
			out[k] = v
			continue
		}

		if inj.opts.DryRun {
			out[k] = fmt.Sprintf("<injected:%s#%s>", path, field)
			res.Injected++
			continue
		}

		secrets, ok := cache[path]
		if !ok {
			var err error
			secrets, err = inj.reader.Read(ctx, path)
			if err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("inject: read %s: %w", path, err))
				out[k] = v
				continue
			}
			cache[path] = secrets
		}

		val, exists := secrets[field]
		if !exists {
			res.Errors = append(res.Errors, fmt.Errorf("inject: field %q not found at %s", field, path))
			out[k] = v
			continue
		}

		out[k] = val
		res.Injected++
	}

	return out, res
}
