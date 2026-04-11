package vault

import (
	"context"
	"fmt"
	"strings"
)

// ReadSecret reads a KV v2 secret at the given path and returns its data as a flat map.
func (c *Client) ReadSecret(ctx context.Context, path string) (map[string]string, error) {
	// KV v2 mounts secrets under "<mount>/data/<path>"
	kvPath := kvV2DataPath(path)

	secret, err := c.vault.KVv2(kvMount(path)).Get(ctx, kvKey(path))
	if err != nil {
		return nil, fmt.Errorf("reading secret %q: %w", path, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret %q not found or empty", kvPath)
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		if sv, ok := v.(string); ok {
			result[k] = sv
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// WriteSecret writes the provided data map to a KV v2 secret at the given path.
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]string) error {
	payload := make(map[string]interface{}, len(data))
	for k, v := range data {
		payload[k] = v
	}

	_, err := c.vault.KVv2(kvMount(path)).Put(ctx, kvKey(path), payload)
	if err != nil {
		return fmt.Errorf("writing secret %q: %w", path, err)
	}
	return nil
}

// kvMount extracts the KV mount from a full path (first segment).
func kvMount(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return "secret"
	}
	return parts[0]
}

// kvKey extracts the secret key from a full path (everything after the mount).
func kvKey(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return path
	}
	return parts[1]
}

// kvV2DataPath returns the internal KV v2 data path for display purposes.
func kvV2DataPath(path string) string {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return path
	}
	return parts[0] + "/data/" + parts[1]
}
