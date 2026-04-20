// Package reorder provides key reordering for Vault secrets.
// Keys within a secret can be sorted alphabetically, by a custom
// order list, or reversed — useful for normalising secret layouts
// across namespaces before diffing or exporting.
package reorder

import (
	"context"
	"fmt"
	"sort"
)

// Reader reads a secret's key/value pairs from a path.
type Reader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes key/value pairs to a path.
type Writer interface {
	WriteSecret(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Options controls how keys are reordered.
type Options struct {
	// Order is an explicit key ordering. Keys not present in Order
	// are appended alphabetically after the ordered keys.
	Order []string
	// Reverse sorts keys in reverse alphabetical order when Order is empty.
	Reverse bool
	// DryRun reports what would change without writing.
	DryRun bool
}

// Delta records a single key whose position changed.
type Delta struct {
	Path string
	Key  string
	From int
	To   int
}

// Result holds the outcome of a reorder operation.
type Result struct {
	Deltas []Delta
	DryRun bool
	Err    error
}

// Reorderer applies key reordering to Vault secret paths.
type Reorderer struct {
	rw   ReadWriter
	opts Options
}

// New creates a new Reorderer.
func New(rw ReadWriter, opts Options) *Reorderer {
	return &Reorderer{rw: rw, opts: opts}
}

// Apply reorders keys at the given path and returns the result.
func (r *Reorderer) Apply(ctx context.Context, path string) Result {
	data, err := r.rw.ReadSecret(ctx, path)
	if err != nil {
		return Result{Err: fmt.Errorf("reorder: read %s: %w", path, err)}
	}

	origOrder := sortedKeys(data)
	newOrder := r.computeOrder(origOrder)

	var deltas []Delta
	for newIdx, key := range newOrder {
		oldIdx := indexOf(origOrder, key)
		if oldIdx != newIdx {
			deltas = append(deltas, Delta{Path: path, Key: key, From: oldIdx, To: newIdx})
		}
	}

	if r.opts.DryRun || len(deltas) == 0 {
		return Result{Deltas: deltas, DryRun: r.opts.DryRun}
	}

	if err := r.rw.WriteSecret(ctx, path, data); err != nil {
		return Result{Deltas: deltas, Err: fmt.Errorf("reorder: write %s: %w", path, err)}
	}
	return Result{Deltas: deltas}
}

func (r *Reorderer) computeOrder(keys []string) []string {
	if len(r.opts.Order) > 0 {
		seen := make(map[string]bool, len(r.opts.Order))
		result := make([]string, 0, len(keys))
		for _, k := range r.opts.Order {
			if _, ok := indexOf2(keys, k); ok {
				result = append(result, k)
				seen[k] = true
			}
		}
		for _, k := range keys {
			if !seen[k] {
				result = append(result, k)
			}
		}
		return result
	}
	out := make([]string, len(keys))
	copy(out, keys)
	if r.opts.Reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(out)))
	} else {
		sort.Strings(out)
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func indexOf(keys []string, key string) int {
	for i, k := range keys {
		if k == key {
			return i
		}
	}
	return -1
}

func indexOf2(keys []string, key string) (int, bool) {
	for i, k := range keys {
		if k == key {
			return i, true
		}
	}
	return -1, false
}
