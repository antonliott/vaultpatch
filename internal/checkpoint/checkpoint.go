// Package checkpoint provides point-in-time labelling of Vault secret paths,
// allowing operators to mark a known-good state and later diff or restore to it.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Checkpoint represents a labelled snapshot of secret paths at a point in time.
type Checkpoint struct {
	Label     string            `json:"label"`
	Namespace string            `json:"namespace"`
	CreatedAt time.Time         `json:"created_at"`
	Paths     []string          `json:"paths"`
	Secrets   map[string]Secret `json:"secrets"`
}

// Secret holds the key-value data captured at checkpoint time.
type Secret struct {
	Data map[string]string `json:"data"`
}

// New creates a new Checkpoint with the given label, namespace, and captured secrets.
func New(label, namespace string, secrets map[string]Secret) *Checkpoint {
	paths := make([]string, 0, len(secrets))
	for p := range secrets {
		paths = append(paths, p)
	}
	return &Checkpoint{
		Label:     label,
		Namespace: namespace,
		CreatedAt: time.Now().UTC(),
		Paths:     paths,
		Secrets:   secrets,
	}
}

// Save writes the checkpoint to a JSON file at the given path.
func Save(c *Checkpoint, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("checkpoint: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("checkpoint: encode: %w", err)
	}
	return nil
}

// Load reads a checkpoint from a JSON file.
func Load(filePath string) (*Checkpoint, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("checkpoint: open file: %w", err)
	}
	defer f.Close()
	var c Checkpoint
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("checkpoint: decode: %w", err)
	}
	return &c, nil
}
