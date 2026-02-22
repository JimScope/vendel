package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

const encryptedPrefix = "enc:"

// deriveKey returns a 32-byte AES key from the WEBHOOK_ENCRYPTION_KEY env var.
// Falls back to FIRST_SUPERUSER_PASSWORD as a default key.
func deriveKey() ([]byte, error) {
	raw := os.Getenv("WEBHOOK_ENCRYPTION_KEY")
	if raw == "" {
		raw = os.Getenv("FIRST_SUPERUSER_PASSWORD")
	}
	if raw == "" {
		return nil, fmt.Errorf("no encryption key configured")
	}
	hash := sha256.Sum256([]byte(raw))
	return hash[:], nil
}

// EncryptSecret encrypts a plaintext secret using AES-GCM.
// Returns a base64-encoded string prefixed with "enc:".
// If the value is already encrypted or empty, it is returned as-is.
func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, encryptedPrefix) {
		return plaintext, nil
	}

	key, err := deriveKey()
	if err != nil {
		return plaintext, nil // no key configured — store plaintext (backwards compatible)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encryptedPrefix + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts an AES-GCM encrypted secret.
// If the value is not prefixed with "enc:", it is returned as-is (plaintext fallback).
func DecryptSecret(encrypted string) (string, error) {
	if encrypted == "" || !strings.HasPrefix(encrypted, encryptedPrefix) {
		return encrypted, nil // plaintext — backwards compatible
	}

	key, err := deriveKey()
	if err != nil {
		return "", fmt.Errorf("no encryption key: %w", err)
	}

	data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(
		strings.TrimPrefix(encrypted, encryptedPrefix),
	)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}
