// Package search provides functionality to search for secrets across
// Vault paths by key name or value pattern.
package search

import (
	"context"
	"regexp"
)

// Reader can list and read secrets from a Vault path.
type Reader interface {
	List(ctx context.Context, path string) ([]string, error)
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Match represents a single search result.
type Match struct {
	Path  string
	Key   string
	Value string
}

// Searcher searches secrets using a compiled pattern.
type Searcher struct {
	reader  Reader
	mount   string
	keyRe   *regexp.Regexp
	valueRe *regexp.Regexp
}

// NewSearcher creates a Searcher. keyPattern and valuePattern may be empty
// strings, in which case the respective filter is skipped.
func NewSearcher(r Reader, mount, keyPattern, valuePattern string) (*Searcher, error) {
	var kre, vre *regexp.Regexp
	var err error
	if keyPattern != "" {
		if kre, err = regexp.Compile(keyPattern); err != nil {
			return nil, err
		}
	}
	if valuePattern != "" {
		if vre, err = regexp.Compile(valuePattern); err != nil {
			return nil, err
		}
	}
	return &Searcher{reader: r, mount: mount, keyRe: kre, valueRe: vre}, nil
}

// matches reports whether the key-value pair satisfies the searcher's filters.
func (s *Searcher) matches(k, v string) bool {
	if s.keyRe != nil && !s.keyRe.MatchString(k) {
		return false
	}
	if s.valueRe != nil && !s.valueRe.MatchString(v) {
		return false
	}
	return true
}

// Run walks all paths under mount and returns matches.
func (s *Searcher) Run(ctx context.Context) ([]Match, error) {
	paths, err := s.reader.List(ctx, s.mount)
	if err != nil {
		return nil, err
	}
	var results []Match
	for _, p := range paths {
		secrets, err := s.reader.Read(ctx, p)
		if err != nil {
			continue
		}
		for k, v := range secrets {
			if s.matches(k, v) {
				results = append(results, Match{Path: p, Key: k, Value: v})
			}
		}
	}
	return results, nil
}
