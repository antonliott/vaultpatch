// Package dedupe identifies and removes duplicate secret values across Vault paths.
package dedupe

import (
	"context"
	"fmt"
	"sort"
)

// Reader lists and reads secrets from a Vault path.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Duplicate holds paths that share identical secret values for a given key.
type Duplicate struct {
	Key   string
	Value string
	Paths []string
}

// Result contains all duplicates found across the scanned paths.
type Result struct {
	Duplicates []Duplicate
}

// HasDuplicates returns true when at least one duplicate was found.
func (r Result) HasDuplicates() bool {
	return len(r.Duplicates) > 0
}

// Summary returns a human-readable summary of the result.
func (r Result) Summary() string {
	if !r.HasDuplicates() {
		return "no duplicate secret values found"
	}
	return fmt.Sprintf("%d duplicate value(s) found across secret paths", len(r.Duplicates))
}

// Finder scans Vault paths for duplicate secret values.
type Finder struct {
	reader Reader
}

// NewFinder returns a Finder backed by the provided Reader.
func NewFinder(r Reader) *Finder {
	return &Finder{reader: r}
}

// Find reads all secrets under mountPath and returns keys whose values appear
// at more than one path.
func (f *Finder) Find(ctx context.Context, mountPath string) (Result, error) {
	paths, err := f.reader.List(ctx, mountPath)
	if err != nil {
		return Result{}, fmt.Errorf("dedupe: list %q: %w", mountPath, err)
	}

	// index: key+"\x00"+value -> []path
	index := make(map[string][]string)

	for _, p := range paths {
		secrets, err := f.reader.Read(ctx, p)
		if err != nil {
			return Result{}, fmt.Errorf("dedupe: read %q: %w", p, err)
		}
		for k, v := range secrets {
			token := k + "\x00" + v
			index[token] = append(index[token], p)
		}
	}

	var dups []Duplicate
	for token, ps := range index {
		if len(ps) < 2 {
			continue
		}
		// split token back into key and value
		var key, val string
		for i, c := range token {
			if c == '\x00' {
				key = token[:i]
				val = token[i+1:]
				break
			}
		}
		sort.Strings(ps)
		dups = append(dups, Duplicate{Key: key, Value: val, Paths: ps})
	}

	sort.Slice(dups, func(i, j int) bool {
		if dups[i].Key != dups[j].Key {
			return dups[i].Key < dups[j].Key
		}
		return dups[i].Value < dups[j].Value
	})

	return Result{Duplicates: dups}, nil
}
