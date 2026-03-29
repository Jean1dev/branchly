package infra

import (
	"strings"
	"testing"
)

var testKey = func() []byte {
	k := make([]byte, 32)
	for i := range k {
		k[i] = byte(i + 1)
	}
	return k
}()

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plain := "ghp_supersecrettoken123"
	enc, err := Encrypt(plain, testKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	got, err := Decrypt(enc, testKey)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Errorf("Decrypt = %q, want %q", got, plain)
	}
}

func TestEncryptProducesUniqueNonces(t *testing.T) {
	plain := "same-plaintext"
	enc1, err := Encrypt(plain, testKey)
	if err != nil {
		t.Fatalf("Encrypt 1: %v", err)
	}
	enc2, err := Encrypt(plain, testKey)
	if err != nil {
		t.Fatalf("Encrypt 2: %v", err)
	}
	if enc1 == enc2 {
		t.Error("two Encrypt calls with same plaintext produced identical ciphertext — nonces must be unique")
	}
}

func TestDecryptWrongKeyReturnsError(t *testing.T) {
	enc, err := Encrypt("secret", testKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	wrongKey := make([]byte, 32)
	_, err = Decrypt(enc, wrongKey)
	if err == nil {
		t.Error("Decrypt with wrong key should return an error")
	}
}

func TestDecryptCorruptedCiphertextReturnsError(t *testing.T) {
	enc, err := Encrypt("secret", testKey)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	corrupted := strings.ReplaceAll(enc, enc[len(enc)-4:], "AAAA")
	_, err = Decrypt(corrupted, testKey)
	if err == nil {
		t.Error("Decrypt with corrupted ciphertext should return an error")
	}
}

func TestDecryptEmptyStringReturnsError(t *testing.T) {
	_, err := Decrypt("", testKey)
	if err == nil {
		t.Error("Decrypt with empty string should return an error")
	}
}
