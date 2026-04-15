// Package encrypt provides field-level encryption and decryption
// for Vault secret values using AES-GCM.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const prefix = "enc:"

// Encrypter encrypts and decrypts secret map values.
type Encrypter struct {
	block cipher.Block
}

// New creates an Encrypter from a 16, 24, or 32-byte AES key.
func New(key []byte) (*Encrypter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encrypt: invalid key: %w", err)
	}
	return &Encrypter{block: block}, nil
}

// Encrypt encrypts all values in the provided map, returning a new map.
// Values that are already encrypted (prefixed with "enc:") are skipped.
func (e *Encrypter) Encrypt(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if len(v) >= len(prefix) && v[:len(prefix)] == prefix {
			out[k] = v
			continue
		}
		ciph, err := e.seal([]byte(v))
		if err != nil {
			return nil, fmt.Errorf("encrypt: key %q: %w", k, err)
		}
		out[k] = prefix + base64.StdEncoding.EncodeToString(ciph)
	}
	return out, nil
}

// Decrypt decrypts all values in the provided map, returning a new map.
// Values without the "enc:" prefix are passed through unchanged.
func (e *Encrypter) Decrypt(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if len(v) < len(prefix) || v[:len(prefix)] != prefix {
			out[k] = v
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(v[len(prefix):])
		if err != nil {
			return nil, fmt.Errorf("encrypt: key %q: base64 decode: %w", k, err)
		}
		plain, err := e.open(raw)
		if err != nil {
			return nil, fmt.Errorf("encrypt: key %q: %w", k, err)
		}
		out[k] = string(plain)
	}
	return out, nil
}

func (e *Encrypter) seal(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (e *Encrypter) open(data []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, errors.New("ciphertext too short")
	}
	return gcm.Open(nil, data[:ns], data[ns:], nil)
}
