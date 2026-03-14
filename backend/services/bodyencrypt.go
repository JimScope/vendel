package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

	return aesGCMEncrypt(key, plaintext, fieldEncPrefix)
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

	return aesGCMDecrypt(key, value, fieldEncPrefix)
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
