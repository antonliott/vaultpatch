package expire

import (
	"fmt"
	"time"
)

// SecretReader lists and reads secrets from a KV path.
type SecretReader interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
}

// SecretWriter writes secrets to a KV path.
type SecretWriter interface {
	Write(path string, data map[string]string) error
}

// Finding describes a secret key that has expired or is expiring soon.
type Finding struct {
	Path    string
	Key     string
	Expires time.Time
	Expired bool
}

// Checker scans secrets for expiry metadata.
type Checker struct {
	reader    SecretReader
	metaKey   string
	warnAfter time.Duration
}

// NewChecker returns a Checker. metaKey is the secret field holding the RFC3339 expiry timestamp.
func NewChecker(r SecretReader, metaKey string, warnAfter time.Duration) *Checker {
	if metaKey == "" {
		metaKey = "expires_at"
	}
	return &Checker{reader: r, metaKey: metaKey, warnAfter: warnAfter}
}

// Scan lists all secrets under path and returns expiry findings.
func (c *Checker) Scan(path string) ([]Finding, error) {
	keys, err := c.reader.List(path)
	if err != nil {
		return nil, fmt.Errorf("expire: list %s: %w", path, err)
	}

	now := time.Now().UTC()
	var findings []Finding

	for _, key := range keys {
		data, err := c.reader.Read(path + "/" + key)
		if err != nil {
			return nil, fmt.Errorf("expire: read %s/%s: %w", path, key, err)
		}
		raw, ok := data[c.metaKey]
		if !ok {
			continue
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			continue
		}
		t = t.UTC()
		if t.Before(now) || t.Before(now.Add(c.warnAfter)) {
			findings = append(findings, Finding{
				Path:    path + "/" + key,
				Key:     c.metaKey,
				Expires: t,
				Expired: t.Before(now),
			})
		}
	}
	return findings, nil
}
