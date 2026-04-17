// Package census provides secret count and key distribution statistics
// across one or more Vault paths.
package census

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

// PathStat holds statistics for a single secret path.
type PathStat struct {
	Path      string
	KeyCount  int
	Keys      []string
}

// Report holds the full census result.
type Report struct {
	TotalPaths  int
	TotalKeys   int
	KeyFreq     map[string]int // how many paths contain each key name
	Paths       []PathStat
}

// Collector gathers census data.
type Collector struct {
	reader Reader
}

// NewCollector returns a new Collector.
func NewCollector(r Reader) *Collector {
	return &Collector{reader: r}
}

// Collect walks the given root path and builds a Report.
func (c *Collector) Collect(ctx context.Context, root string) (*Report, error) {
	paths, err := c.reader.List(ctx, root)
	if err != nil {
		return nil, fmt.Errorf("census: list %q: %w", root, err)
	}

	report := &Report{
		KeyFreq: make(map[string]int),
	}

	for _, p := range paths {
		full := root + "/" + p
		secrets, err := c.reader.Read(ctx, full)
		if err != nil {
			return nil, fmt.Errorf("census: read %q: %w", full, err)
		}

		stat := PathStat{Path: full}
		for k := range secrets {
			stat.Keys = append(stat.Keys, k)
			report.KeyFreq[k]++
		}
		sort.Strings(stat.Keys)
		stat.KeyCount = len(stat.Keys)

		report.Paths = append(report.Paths, stat)
		report.TotalKeys += stat.KeyCount
	}

	report.TotalPaths = len(report.Paths)
	return report, nil
}
