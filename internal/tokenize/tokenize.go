// Package tokenize provides utilities for replacing secret values in
// arbitrary text with opaque placeholder tokens, and reversing the
// substitution when the original values are needed again.
package tokenize

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// Tokenizer replaces secret values with random tokens and can reverse
// the substitution via Detokenize.
type Tokenizer struct {
	mu      sync.RWMutex
	forward map[string]string // value -> token
	reverse map[string]string // token -> value
	prefix  string
}

// New returns a Tokenizer whose generated tokens are prefixed with
// prefix (e.g. "VAULT"). An empty prefix defaults to "TOKEN".
func New(prefix string) *Tokenizer {
	if prefix == "" {
		prefix = "TOKEN"
	}
	return &Tokenizer{
		forward: make(map[string]string),
		reverse: make(map[string]string),
		prefix:  prefix,
	}
}

// Token returns a stable opaque token for value. Calling Token with
// the same value always returns the same token within a Tokenizer's
// lifetime.
func (t *Tokenizer) Token(value string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if tok, ok := t.forward[value]; ok {
		return tok, nil
	}

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("tokenize: generate token: %w", err)
	}
	tok := fmt.Sprintf("<%s_%s>", t.prefix, hex.EncodeToString(b))
	t.forward[value] = tok
	t.reverse[tok] = value
	return tok, nil
}

// Apply replaces every map value with its corresponding token.
// The original map is not mutated; a new map is returned.
func (t *Tokenizer) Apply(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		tok, err := t.Token(v)
		if err != nil {
			return nil, err
		}
		out[k] = tok
	}
	return out, nil
}

// Detokenize replaces all known tokens embedded in text with their
// original values. Tokens that were never registered are left as-is.
func (t *Tokenizer) Detokenize(text string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for tok, val := range t.reverse {
		text = strings.ReplaceAll(text, tok, val)
	}
	return text
}

// Len returns the number of distinct values currently registered.
func (t *Tokenizer) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.forward)
}
