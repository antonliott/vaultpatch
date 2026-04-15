// Package digest computes and compares content hashes for Vault secret paths,
// enabling change detection without storing full secret values.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Reader can list and read secrets from a Vault path.
type Reader interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
}

// Entry holds the digest for a single secret path.
type Entry struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// Result maps secret paths to their computed digest entries.
type Result map[string]Entry

// Compute walks all secrets under mountPath and returns a digest per path.
func Compute(r Reader, mountPath string) (Result, error) {
	keys, err := r.List(mountPath)
	if err != nil {
		return nil, fmt.Errorf("digest: list %q: %w", mountPath, err)
	}

	out := make(Result, len(keys))
	for _, k := range keys {
		full := strings.TrimRight(mountPath, "/") + "/" + k
		secrets, err := r.Read(full)
		if err != nil {
			return nil, fmt.Errorf("digest: read %q: %w", full, err)
		}
		h := hashMap(secrets)
		out[full] = Entry{Path: full, Digest: h}
	}
	return out, nil
}

// Changed returns paths whose digest differs between a and b, plus paths
// present in only one of the two results.
func Changed(a, b Result) []string {
	seen := make(map[string]struct{})
	var diffs []string

	for path, ea := range a {
		seen[path] = struct{}{}
		if eb, ok := b[path]; !ok || ea.Digest != eb.Digest {
			diffs = append(diffs, path)
		}
	}
	for path := range b {
		if _, ok := seen[path]; !ok {
			diffs = append(diffs, path)
		}
	}
	sort.Strings(diffs)
	return diffs
}

// hashMap produces a stable SHA-256 hex digest of a key-value map.
func hashMap(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%s=%s\n", k, m[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}
