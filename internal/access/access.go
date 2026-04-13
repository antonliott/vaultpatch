// Package access provides functionality for auditing and comparing
// Vault token accessor metadata across namespaces.
package access

import (
	"context"
	"fmt"
	"sort"
)

// TokenInfo holds metadata about a Vault token accessor.
type TokenInfo struct {
	Accessor   string            `json:"accessor"`
	DisplayName string           `json:"display_name"`
	Policies   []string          `json:"policies"`
	Meta       map[string]string `json:"meta,omitempty"`
	TTL        int               `json:"ttl"`
}

// Delta represents a difference in token accessor metadata.
type Delta struct {
	Accessor string
	Field    string
	Old      string
	New      string
}

// TokenReader is the interface for listing and looking up token accessors.
type TokenReader interface {
	ListAccessors(ctx context.Context) ([]string, error)
	LookupAccessor(ctx context.Context, accessor string) (*TokenInfo, error)
}

// Auditor inspects token accessors and computes deltas.
type Auditor struct {
	reader TokenReader
}

// NewAuditor creates a new Auditor backed by the given TokenReader.
func NewAuditor(r TokenReader) *Auditor {
	return &Auditor{reader: r}
}

// Collect returns all TokenInfo entries for the current namespace.
func (a *Auditor) Collect(ctx context.Context) ([]*TokenInfo, error) {
	accessors, err := a.reader.ListAccessors(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accessors: %w", err)
	}

	var infos []*TokenInfo
	for _, acc := range accessors {
		info, err := a.reader.LookupAccessor(ctx, acc)
		if err != nil {
			return nil, fmt.Errorf("lookup accessor %s: %w", acc, err)
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// Compare returns deltas between two sets of TokenInfo keyed by accessor.
func Compare(src, dst []*TokenInfo) []Delta {
	srcMap := make(map[string]*TokenInfo, len(src))
	for _, t := range src {
		srcMap[t.Accessor] = t
	}

	var deltas []Delta
	for _, d := range dst {
		s, ok := srcMap[d.Accessor]
		if !ok {
			continue
		}
		if s.DisplayName != d.DisplayName {
			deltas = append(deltas, Delta{Accessor: d.Accessor, Field: "display_name", Old: s.DisplayName, New: d.DisplayName})
		}
		if s.TTL != d.TTL {
			deltas = append(deltas, Delta{Accessor: d.Accessor, Field: "ttl", Old: fmt.Sprintf("%d", s.TTL), New: fmt.Sprintf("%d", d.TTL)})
		}
		sortedSrc := sortedPolicies(s.Policies)
		sortedDst := sortedPolicies(d.Policies)
		if sortedSrc != sortedDst {
			deltas = append(deltas, Delta{Accessor: d.Accessor, Field: "policies", Old: sortedSrc, New: sortedDst})
		}
	}
	return deltas
}

func sortedPolicies(p []string) string {
	cp := make([]string, len(p))
	copy(cp, p)
	sort.Strings(cp)
	result := ""
	for i, v := range cp {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}
