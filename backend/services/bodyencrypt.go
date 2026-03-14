package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

const fieldEncPrefix = "fenc:"

// deriveKeyWithPurpose derives a 32-byte AES key using HMAC-SHA256
// with a purpose string, providing cryptographic key separation from
// the same WEBHOOK_ENCRYPTION_KEY master secret.
func deriveKeyWithPurpose(purpose string) ([]byte, error) {
	raw := os.Getenv("WEBHOOK_ENCRYPTION_KEY")
	if raw == "" {
		return nil, fmt.Errorf("WEBHOOK_ENCRYPTION_KEY not set")
	}
	mac := hmac.New(sha256.New, []byte(raw))
	mac.Write([]byte(purpose))
	return mac.Sum(nil), nil
}

// EncryptBody encrypts an SMS body using AES-GCM with a purpose-derived key.
// Returns a string prefixed with "fenc:". Already-encrypted or empty values pass through.
func EncryptBody(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, fieldEncPrefix) {
		return plaintext, nil
	}

	key, err := deriveKeyWithPurpose("sms-body")
	if err != nil {
		return "", fmt.Errorf("encrypt body: %w", err)
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
	return fieldEncPrefix + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext), nil
}

// DecryptBody decrypts an AES-GCM encrypted body.
// Plaintext values (not prefixed with "fenc:") are returned as-is for backwards compatibility.
func DecryptBody(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, fieldEncPrefix) {
		return value, nil
	}

	key, err := deriveKeyWithPurpose("sms-body")
	if err != nil {
		return "", fmt.Errorf("decrypt body: %w", err)
	}

	data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(
		strings.TrimPrefix(value, fieldEncPrefix),
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

// ComputeBodyHash computes a deterministic HMAC-SHA256 hash of the body
// for use as a blind index in deduplication queries.
func ComputeBodyHash(plaintext string) (string, error) {
	key, err := deriveKeyWithPurpose("sms-body-hash")
	if err != nil {
		return "", fmt.Errorf("compute body hash: %w", err)
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// IsBodyEncrypted returns true if the value has the field encryption prefix.
func IsBodyEncrypted(value string) bool {
	return strings.HasPrefix(value, fieldEncPrefix)
}

// GetRecordBody returns the decrypted body from a record.
// Falls back to raw value if decryption fails (e.g. pre-migration plaintext).
func GetRecordBody(record *core.Record) string {
	raw := record.GetString("body")
	decrypted, err := DecryptBody(raw)
	if err != nil {
		return raw
	}
	return decrypted
}
