package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time capture of Vault secrets.
type Snapshot struct {
	Namespace string            `json:"namespace"`
	CapturedAt time.Time        `json:"captured_at"`
	Secrets    map[string]string `json:"secrets"`
}

// New creates a new Snapshot for the given namespace and secrets map.
func New(namespace string, secrets map[string]string) *Snapshot {
	return &Snapshot{
		Namespace:  namespace,
		CapturedAt: time.Now().UTC(),
		Secrets:    secrets,
	}
}

// Save writes the snapshot as JSON to the given file path.
func Save(s *Snapshot, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("snapshot: create file %q: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// Load reads a snapshot from the given file path.
func Load(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open file %q: %w", path, err)
	}
	defer f.Close()

	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}
	return &s, nil
}
