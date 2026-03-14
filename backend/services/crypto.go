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

// GenerateKeyPrefix returns the first 10 characters of a key followed by "...".
func GenerateKeyPrefix(key string) string {
	if len(key) <= 10 {
		return key
	}
	return key[:10] + "..."
}
