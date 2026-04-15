package checkpoint

import (
	"context"
	"fmt"
)

// Reader can list and read secrets from a Vault path.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Collector gathers secrets from Vault into a Checkpoint.
type Collector struct {
	reader    Reader
	namespace string
}

// NewCollector returns a Collector backed by the given Reader.
func NewCollector(r Reader, namespace string) *Collector {
	return &Collector{reader: r, namespace: namespace}
}

// Collect reads all secrets under the given paths and returns a map suitable
// for constructing a Checkpoint.
func (c *Collector) Collect(ctx context.Context, paths []string) (map[string]Secret, error) {
	result := make(map[string]Secret, len(paths))
	for _, p := range paths {
		keys, err := c.reader.List(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("checkpoint: list %q: %w", p, err)
		}
		for _, k := range keys {
			full := p + "/" + k
			data, err := c.reader.Read(ctx, full)
			if err != nil {
				return nil, fmt.Errorf("checkpoint: read %q: %w", full, err)
			}
			result[full] = Secret{Data: data}
		}
	}
	return result, nil
}
