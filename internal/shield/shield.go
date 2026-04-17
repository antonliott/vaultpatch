// Package shield provides write-protection for Vault secret paths.
package shield

import (
	"fmt"
	"sort"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	Read(path string) (map[string]string, error)
}

// Writer writes secrets to a Vault path.
type Writer interface {
	Write(path string, data map[string]string) error
}

// Result holds the outcome of a shield check.
type Result struct {
	Path      string
	Protected []string
	Blocked   bool
}

// Shielder enforces write-protection on specified keys.
type Shielder struct {
	rw      interface {
		Reader
		Writer
	}
	protect map[string]struct{}
	dryRun  bool
}

// NewShielder creates a Shielder that protects the given keys.
func NewShielder(rw interface {
	Reader
	Writer
}, keys []string, dryRun bool) *Shielder {
	p := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		p[k] = struct{}{}
	}
	return &Shielder{rw: rw, protect: p, dryRun: dryRun}
}

// Apply writes data to path, stripping any protected keys from the update.
// Returns a Result describing which keys were blocked.
func (s *Shielder) Apply(path string, incoming map[string]string) (Result, error) {
	existing, err := s.rw.Read(path)
	if err != nil {
		return Result{}, fmt.Errorf("shield: read %s: %w", path, err)
	}

	merged := make(map[string]string, len(existing))
	for k, v := range existing {
		merged[k] = v
	}

	var blocked []string
	for k, v := range incoming {
		if _, protected := s.protect[k]; protected {
			blocked = append(blocked, k)
			continue
		}
		merged[k] = v
	}
	sort.Strings(blocked)

	res := Result{Path: path, Protected: blocked, Blocked: len(blocked) > 0}
	if s.dryRun {
		return res, nil
	}
	if err := s.rw.Write(path, merged); err != nil {
		return res, fmt.Errorf("shield: write %s: %w", path, err)
	}
	return res, nil
}
