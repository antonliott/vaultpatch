// Package summarize provides a high-level summary of secrets stored at a
// Vault path: total key count, value length statistics, and type hints.
package summarize

import (
	"fmt"
	"sort"
	"strings"
)

// Reader can list and read KV secrets.
type Reader interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
}

// KeyStat holds per-key statistics.
type KeyStat struct {
	Key       string
	ValueLen  int
	TypeHint  string
}

// Summary holds aggregate statistics for a path.
type Summary struct {
	Path      string
	TotalKeys int
	AvgLen    float64
	MaxLen    int
	MinLen    int
	Stats     []KeyStat
}

// Summarizer computes summaries from Vault paths.
type Summarizer struct {
	reader Reader
}

// New returns a new Summarizer.
func New(r Reader) *Summarizer {
	return &Summarizer{reader: r}
}

// Compute reads the secret at path and returns a Summary.
func (s *Summarizer) Compute(path string) (*Summary, error) {
	data, err := s.reader.Read(path)
	if err != nil {
		return nil, fmt.Errorf("summarize: read %q: %w", path, err)
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sum := &Summary{
		Path:      path,
		TotalKeys: len(keys),
		MinLen:    -1,
	}

	var total int
	for _, k := range keys {
		v := data[k]
		vl := len(v)
		total += vl
		if vl > sum.MaxLen {
			sum.MaxLen = vl
		}
		if sum.MinLen == -1 || vl < sum.MinLen {
			sum.MinLen = vl
		}
		sum.Stats = append(sum.Stats, KeyStat{
			Key:      k,
			ValueLen: vl,
			TypeHint: typeHint(v),
		})
	}

	if len(keys) > 0 {
		sum.AvgLen = float64(total) / float64(len(keys))
	} else {
		sum.MinLen = 0
	}
	return sum, nil
}

func typeHint(v string) string {
	switch {
	case strings.HasPrefix(v, "vault:v"):
		return "encrypted"
	case len(v) == 36 && strings.Count(v, "-") == 4:
		return "uuid"
	case strings.Contains(v, "://"):
		return "url"
	default:
		return "string"
	}
}
