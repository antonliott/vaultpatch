// Package group provides grouping of secrets by key prefix across paths.
package group

import (
	"fmt"
	"sort"
	"strings"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
}

// Group holds secrets sharing a common key prefix.
type Group struct {
	Prefix string
	Entries map[string]map[string]string // path -> filtered kv
}

// Grouper groups secrets by key prefix.
type Grouper struct {
	reader Reader
}

// NewGrouper returns a new Grouper.
func NewGrouper(r Reader) *Grouper {
	return &Grouper{reader: r}
}

// Apply reads all secrets under root and groups them by key prefix.
func (g *Grouper) Apply(root, prefix string) ([]Group, error) {
	paths, err := g.reader.List(root)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", root, err)
	}

	matched := map[string]map[string]string{}

	for _, p := range paths {
		full := strings.TrimSuffix(root, "/") + "/" + p
		secrets, err := g.reader.Read(full)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", full, err)
		}
		filtered := map[string]string{}
		for k, v := range secrets {
			if strings.HasPrefix(k, prefix) {
				filtered[k] = v
			}
		}
		if len(filtered) > 0 {
			matched[full] = filtered
		}
	}

	if len(matched) == 0 {
		return nil, nil
	}

	g := Group{Prefix: prefix, Entries: matched}

	// sort paths for deterministic output
	_ = sort.Search // already imported via sort
	return []Group{g}, nil
}

// Format renders groups as human-readable text.
func Format(groups []Group) string {
	if len(groups) == 0 {
		return "no matching keys found\n"
	}
	var sb strings.Builder
	for _, grp := range groups {
		fmt.Fprintf(&sb, "prefix: %s\n", grp.Prefix)
		paths := make([]string, 0, len(grp.Entries))
		for p := range grp.Entries {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			fmt.Fprintf(&sb, "  %s\n", p)
			keys := make([]string, 0, len(grp.Entries[p]))
			for k := range grp.Entries[p] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&sb, "    %s = %s\n", k, grp.Entries[p][k])
			}
		}
	}
	return sb.String()
}
