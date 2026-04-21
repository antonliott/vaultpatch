// Package linkage detects and reports cross-path secret key references
// within a Vault namespace, identifying keys whose values match paths
// or key patterns in other secrets.
package linkage

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Reader can list and read KV secrets.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Link describes a detected reference between two secret paths.
type Link struct {
	SourcePath string
	SourceKey  string
	TargetPath string
}

// Detector scans secrets for cross-path references.
type Detector struct {
	reader Reader
	prefix string
}

// NewDetector returns a Detector scoped to the given mount prefix.
func NewDetector(r Reader, prefix string) *Detector {
	return &Detector{reader: r, prefix: strings.TrimSuffix(prefix, "/")}
}

// Detect lists all secrets under the prefix and returns Links where a
// secret value looks like a path to another known secret.
func (d *Detector) Detect(ctx context.Context) ([]Link, error) {
	paths, err := d.reader.List(ctx, d.prefix)
	if err != nil {
		return nil, fmt.Errorf("linkage: list %q: %w", d.prefix, err)
	}

	// Build a set of known paths for O(1) lookup.
	known := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		known[d.prefix+"/"+strings.TrimSuffix(p, "/")] = struct{}{}
	}

	var links []Link
	for _, p := range paths {
		full := d.prefix + "/" + strings.TrimSuffix(p, "/")
		secret, err := d.reader.Read(ctx, full)
		if err != nil {
			return nil, fmt.Errorf("linkage: read %q: %w", full, err)
		}
		for k, v := range secret {
			v = strings.TrimSpace(v)
			if _, ok := known[v]; ok && v != full {
				links = append(links, Link{
					SourcePath: full,
					SourceKey:  k,
					TargetPath: v,
				})
			}
		}
	}

	sort.Slice(links, func(i, j int) bool {
		if links[i].SourcePath != links[j].SourcePath {
			return links[i].SourcePath < links[j].SourcePath
		}
		return links[i].SourceKey < links[j].SourceKey
	})
	return links, nil
}

// Format renders detected links as a human-readable report.
func Format(links []Link) string {
	if len(links) == 0 {
		return "no cross-path links detected\n"
	}
	var sb strings.Builder
	for _, l := range links {
		fmt.Fprintf(&sb, "%s[%s] -> %s\n", l.SourcePath, l.SourceKey, l.TargetPath)
	}
	return sb.String()
}
