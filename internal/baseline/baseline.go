// Package baseline captures and compares a known-good state of secrets
// across one or more Vault paths, enabling drift detection against a fixed reference.
package baseline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Reader lists and reads secrets from Vault.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Baseline holds a captured snapshot of secrets at a point in time.
type Baseline struct {
	CapturedAt time.Time                       `json:"captured_at"`
	Namespace  string                          `json:"namespace"`
	Secrets    map[string]map[string]string    `json:"secrets"`
}

// Delta describes the drift between a baseline and current state.
type Delta struct {
	Path    string
	Key     string
	Status  string // "added", "removed", "changed"
	Baseline string
	Current  string
}

// Capture reads all secrets under path and returns a Baseline.
func Capture(ctx context.Context, r Reader, namespace, path string) (*Baseline, error) {
	keys, err := r.List(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("baseline: list %q: %w", path, err)
	}

	secrets := make(map[string]map[string]string, len(keys))
	for _, k := range keys {
		data, err := r.Read(ctx, path+"/"+k)
		if err != nil {
			return nil, fmt.Errorf("baseline: read %q: %w", k, err)
		}
		secrets[k] = data
	}

	return &Baseline{
		CapturedAt: time.Now().UTC(),
		Namespace:  namespace,
		Secrets:    secrets,
	}, nil
}

// Save writes the baseline to a JSON file at dest.
func Save(b *Baseline, dest string) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("baseline: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(b); err != nil {
		return fmt.Errorf("baseline: encode: %w", err)
	}
	return nil
}

// Load reads a baseline from a JSON file at src.
func Load(src string) (*Baseline, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("baseline: open file: %w", err)
	}
	defer f.Close()
	var b Baseline
	if err := json.NewDecoder(f).Decode(&b); err != nil {
		return nil, fmt.Errorf("baseline: decode: %w", err)
	}
	return &b, nil
}

// Compare returns deltas between a saved baseline and the current secrets.
func Compare(ctx context.Context, r Reader, b *Baseline, path string) ([]Delta, error) {
	keys, err := r.List(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("baseline: compare list: %w", err)
	}

	current := make(map[string]map[string]string, len(keys))
	for _, k := range keys {
		data, err := r.Read(ctx, path+"/"+k)
		if err != nil {
			return nil, fmt.Errorf("baseline: compare read %q: %w", k, err)
		}
		current[k] = data
	}

	var deltas []Delta

	for secretKey, baseVals := range b.Secrets {
		curVals, exists := current[secretKey]
		if !exists {
			for k, v := range baseVals {
				deltas = append(deltas, Delta{Path: secretKey, Key: k, Status: "removed", Baseline: v})
			}
			continue
		}
		for k, bv := range baseVals {
			cv, ok := curVals[k]
			if !ok {
				deltas = append(deltas, Delta{Path: secretKey, Key: k, Status: "removed", Baseline: bv})
			} else if cv != bv {
				deltas = append(deltas, Delta{Path: secretKey, Key: k, Status: "changed", Baseline: bv, Current: cv})
			}
		}
		for k, cv := range curVals {
			if _, ok := baseVals[k]; !ok {
				deltas = append(deltas, Delta{Path: secretKey, Key: k, Status: "added", Current: cv})
			}
		}
	}

	for secretKey, curVals := range current {
		if _, ok := b.Secrets[secretKey]; !ok {
			for k, cv := range curVals {
				deltas = append(deltas, Delta{Path: secretKey, Key: k, Status: "added", Current: cv})
			}
		}
	}

	return deltas, nil
}
