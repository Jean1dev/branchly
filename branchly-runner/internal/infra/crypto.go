package infra

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

func ParseEncryptionKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if len(s) == 64 {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decode hex ENCRYPTION_KEY: %w", err)
		}
		if len(b) != 32 {
			return nil, fmt.Errorf("decoded ENCRYPTION_KEY must be 32 bytes")
		}
		return b, nil
	}
	if len(s) == 32 {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("ENCRYPTION_KEY must be 32 raw bytes or 64 hex characters")
}

func Decrypt(ciphertextB64 string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("infra/crypto: key must be 32 bytes")
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("infra/crypto: decode base64: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("infra/crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("infra/crypto: gcm: %w", err)
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("infra/crypto: ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("infra/crypto: decrypt: %w", err)
	}
	return string(plain), nil
}
