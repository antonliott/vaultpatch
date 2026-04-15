package encrypt

import (
	"strings"
	"testing"
)

var testKey = []byte("0123456789abcdef") // 16-byte AES-128 key

func TestNew_ValidKey(t *testing.T) {
	_, err := New(testKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNew_InvalidKey(t *testing.T) {
	_, err := New([]byte("short"))
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	e, _ := New(testKey)
	input := map[string]string{
		"db_password": "s3cr3t",
		"api_key":     "abc123",
	}
	enc, err := e.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	for k, v := range enc {
		if !strings.HasPrefix(v, prefix) {
			t.Errorf("key %q: expected enc: prefix, got %q", k, v)
		}
	}
	dec, err := e.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	for k, want := range input {
		if got := dec[k]; got != want {
			t.Errorf("key %q: got %q, want %q", k, got, want)
		}
	}
}

func TestEncrypt_SkipsAlreadyEncrypted(t *testing.T) {
	e, _ := New(testKey)
	alreadyEnc := map[string]string{
		"token": "enc:aGVsbG8=",
	}
	out, err := e.Encrypt(alreadyEnc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["token"] != alreadyEnc["token"] {
		t.Errorf("expected value unchanged, got %q", out["token"])
	}
}

func TestDecrypt_PassthroughPlaintext(t *testing.T) {
	e, _ := New(testKey)
	input := map[string]string{
		"plain": "not-encrypted",
	}
	out, err := e.Decrypt(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["plain"] != "not-encrypted" {
		t.Errorf("expected passthrough, got %q", out["plain"])
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	e, _ := New(testKey)
	input := map[string]string{
		"bad": "enc:!!!notbase64!!!",
	}
	_, err := e.Decrypt(input)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	e, _ := New(testKey)
	enc, _ := e.Encrypt(map[string]string{"k": "value"})
	enc["k"] = enc["k"] + "tampered"
	_, err := e.Decrypt(enc)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestEncrypt_DoesNotMutateInput(t *testing.T) {
	e, _ := New(testKey)
	input := map[string]string{"pw": "original"}
	_, _ = e.Encrypt(input)
	if input["pw"] != "original" {
		t.Error("Encrypt mutated the input map")
	}
}
