// Package compare provides cross-namespace secret comparison for Vault paths.
package compare

import (
	"context"
	"fmt"
	"sort"
)

// SecretReader reads secrets from a given path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// Delta represents a single key-level difference between two namespaces.
type Delta struct {
	Path      string
	Key       string
	SourceVal interface{}
	TargetVal interface{}
	Kind      DeltaKind
}

// DeltaKind classifies the type of change.
type DeltaKind string

const (
	Added   DeltaKind = "added"
	Removed DeltaKind = "removed"
	Changed DeltaKind = "changed"
)

// Result holds all deltas found between source and target.
type Result struct {
	Deltas []Delta
}

// HasDiff returns true if any deltas exist.
func (r Result) HasDiff() bool { return len(r.Deltas) > 0 }

// NewComparator creates a Comparator using the provided readers.
func NewComparator(src, tgt SecretReader) *Comparator {
	return &Comparator{src: src, tgt: tgt}
}

// Comparator compares secrets between two Vault namespaces.
type Comparator struct {
	src SecretReader
	tgt SecretReader
}

// Compare lists secrets under path in src and diffs each against tgt.
func (c *Comparator) Compare(ctx context.Context, path string) (Result, error) {
	keys, err := c.src.ListSecrets(ctx, path)
	if err != nil {
		return Result{}, fmt.Errorf("list source %q: %w", path, err)
	}

	var deltas []Delta
	for _, key := range keys {
		full := path + "/" + key
		srcData, err := c.src.ReadSecret(ctx, full)
		if err != nil {
			return Result{}, fmt.Errorf("read source %q: %w", full, err)
		}
		tgtData, _ := c.tgt.ReadSecret(ctx, full)

		deltas = append(deltas, diffMaps(full, srcData, tgtData)...)
	}

	sort.Slice(deltas, func(i, j int) bool {
		if deltas[i].Path != deltas[j].Path {
			return deltas[i].Path < deltas[j].Path
		}
		return deltas[i].Key < deltas[j].Key
	})
	return Result{Deltas: deltas}, nil
}

func diffMaps(path string, src, tgt map[string]interface{}) []Delta {
	var out []Delta
	for k, sv := range src {
		tv, ok := tgt[k]
		if !ok {
			out = append(out, Delta{Path: path, Key: k, SourceVal: sv, Kind: Added})
		} else if fmt.Sprintf("%v", sv) != fmt.Sprintf("%v", tv) {
			out = append(out, Delta{Path: path, Key: k, SourceVal: sv, TargetVal: tv, Kind: Changed})
		}
	}
	for k, tv := range tgt {
		if _, ok := src[k]; !ok {
			out = append(out, Delta{Path: path, Key: k, TargetVal: tv, Kind: Removed})
		}
	}
	return out
}
