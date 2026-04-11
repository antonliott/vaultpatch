package snapshot

import (
	"context"
	"fmt"
)

// SecretReader is the interface for reading secrets from Vault.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// Collector gathers secrets from Vault into a Snapshot.
type Collector struct {
	reader    SecretReader
	namespace string
}

// NewCollector creates a Collector for the given namespace and reader.
func NewCollector(namespace string, reader SecretReader) *Collector {
	return &Collector{namespace: namespace, reader: reader}
}

// Collect lists and reads all secrets under the given mount path,
// returning a Snapshot of the flattened key-value pairs.
func (c *Collector) Collect(ctx context.Context, mountPath string) (*Snapshot, error) {
	paths, err := c.reader.ListSecrets(ctx, mountPath)
	if err != nil {
		return nil, fmt.Errorf("collector: list secrets at %q: %w", mountPath, err)
	}

	all := make(map[string]string)
	for _, p := range paths {
		data, err := c.reader.ReadSecret(ctx, p)
		if err != nil {
			return nil, fmt.Errorf("collector: read secret %q: %w", p, err)
		}
		for k, v := range data {
			key := p + "/" + k
			all[key] = v
		}
	}

	return New(c.namespace, all), nil
}
