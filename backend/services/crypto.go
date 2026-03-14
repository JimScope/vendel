package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/pocketbase/pocketbase/tools/security"
)

const encryptedPrefix = "enc:"

// deriveKey returns a 32-byte AES key from the WEBHOOK_ENCRYPTION_KEY env var.
func deriveKey() ([]byte, error) {
	raw := os.Getenv("WEBHOOK_ENCRYPTION_KEY")
	if raw == "" {
		return nil, fmt.Errorf("WEBHOOK_ENCRYPTION_KEY not set: webhook secrets will not be encrypted")
	}
	hash := sha256.Sum256([]byte(raw))
	return hash[:], nil
}

// aesGCMEncrypt encrypts plaintext using AES-GCM with the given key,
// returning prefix + base64url(nonce || ciphertext).
func aesGCMEncrypt(key []byte, plaintext, prefix string) (string, error) {
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
	return prefix + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext), nil
}

// aesGCMDecrypt decrypts a prefix + base64url(nonce || ciphertext) string.
func aesGCMDecrypt(key []byte, value, prefix string) (string, error) {
	data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(
		strings.TrimPrefix(value, prefix),
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

// EncryptSecret encrypts a plaintext secret using AES-GCM.
// Returns a base64-encoded string prefixed with "enc:".
// If the value is already encrypted or empty, it is returned as-is.
func EncryptSecret(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, encryptedPrefix) {
		return plaintext, nil
	}

	key, err := deriveKey()
	if err != nil {
		return "", fmt.Errorf("cannot encrypt webhook secret: %w", err)
	}

	return aesGCMEncrypt(key, plaintext, encryptedPrefix)
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

	return aesGCMDecrypt(key, encrypted, encryptedPrefix)
}

// GenerateCallbackState creates an HMAC-signed state token encoding the userId
// and current timestamp. Format: base64url(userId:timestamp:signature).
func GenerateCallbackState(userId string) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", fmt.Errorf("generate callback state: %w", err)
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	data := userId + ":" + ts

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := data + ":" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(token)), nil
}

// VerifyCallbackState verifies a state token and returns the embedded userId.
// Returns an error if the token is invalid, expired, or tampered with.
func VerifyCallbackState(token string, maxAge time.Duration) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", fmt.Errorf("verify callback state: %w", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid state token encoding")
	}

	parts := strings.SplitN(string(raw), ":", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid state token format")
	}

	userId, tsStr, sigStr := parts[0], parts[1], parts[2]

	// Verify signature
	data := userId + ":" + tsStr
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigStr), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid state token signature")
	}

	// Verify expiry
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid state token timestamp")
	}
	if time.Since(time.Unix(ts, 0)) > maxAge {
		return "", fmt.Errorf("state token expired")
	}

	return userId, nil
}

// GenerateSecureKey generates a cryptographically secure random key with a prefix.
func GenerateSecureKey(prefix string, length int) string {
	return prefix + security.RandomString(length)
}

// GenerateKeyPrefix returns the first KeyPrefixDisplay characters of a key followed by "...".
func GenerateKeyPrefix(key string) string {
	if len(key) <= KeyPrefixDisplay {
		return key
	}
	return key[:KeyPrefixDisplay] + "..."
}

// RotateAPIKeyResult holds the output of a successful key rotation.
type RotateAPIKeyResult struct {
	NewKey *core.Record
}

// RotateAPIKey deactivates the old key and creates a new one.
// The new key record has "key" unhidden so the caller can read the raw value once.
func RotateAPIKey(app core.App, userId, keyId, expiresAt string) (*RotateAPIKeyResult, error) {
	oldKey, err := app.FindRecordById("api_keys", keyId)
	if err != nil {
		return nil, fmt.Errorf("API key not found")
	}
	if oldKey.GetString("user") != userId {
		return nil, fmt.Errorf("not your API key")
	}
	if !oldKey.GetBool("is_active") {
		return nil, fmt.Errorf("cannot rotate a revoked key")
	}

	// Deactivate old key
	oldKey.Set("is_active", false)
	if err := app.Save(oldKey); err != nil {
		return nil, fmt.Errorf("failed to deactivate old key: %w", err)
	}

	// Create new key (the OnRecordCreate hook generates the actual key value)
	col, err := app.FindCollectionByNameOrId("api_keys")
	if err != nil {
		return nil, fmt.Errorf("api_keys collection not found: %w", err)
	}

	newKey := core.NewRecord(col)
	newKey.Set("name", oldKey.GetString("name")+" (rotated)")
	newKey.Set("user", userId)
	newKey.Set("is_active", true)
	if expiresAt != "" {
		newKey.Set("expires_at", expiresAt)
	}

	if err := app.Save(newKey); err != nil {
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	// Unhide key so the caller can read the raw value (shown once)
	newKey.Unhide("key")

	return &RotateAPIKeyResult{NewKey: newKey}, nil
}
