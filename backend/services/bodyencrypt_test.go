package services

import (
	"os"
	"strings"
	"testing"
)

func TestEncryptDecryptBody(t *testing.T) {
	os.Setenv("WEBHOOK_ENCRYPTION_KEY", "test-body-encryption-key")
	defer os.Unsetenv("WEBHOOK_ENCRYPTION_KEY")

	t.Run("round-trip", func(t *testing.T) {
		encrypted, err := EncryptBody("Hello SMS body")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(encrypted, fieldEncPrefix) {
			t.Errorf("expected prefix %q", fieldEncPrefix)
		}

		decrypted, err := DecryptBody(encrypted)
		if err != nil {
			t.Fatal(err)
		}
		if decrypted != "Hello SMS body" {
			t.Errorf("got %q, want %q", decrypted, "Hello SMS body")
		}
	})

	t.Run("empty passthrough", func(t *testing.T) {
		enc, _ := EncryptBody("")
		if enc != "" {
			t.Error("empty should pass through")
		}

		dec, _ := DecryptBody("")
		if dec != "" {
			t.Error("empty should pass through on decrypt")
		}
	})

	t.Run("already encrypted passthrough", func(t *testing.T) {
		enc, _ := EncryptBody("fenc:already-done")
		if enc != "fenc:already-done" {
			t.Error("already encrypted should pass through")
		}
	})

	t.Run("plaintext fallback on decrypt", func(t *testing.T) {
		dec, _ := DecryptBody("plain text message")
		if dec != "plain text message" {
			t.Error("non-prefixed values should return as-is")
		}
	})
}

func TestComputeBodyHash(t *testing.T) {
	os.Setenv("WEBHOOK_ENCRYPTION_KEY", "test-hash-key")
	defer os.Unsetenv("WEBHOOK_ENCRYPTION_KEY")

	t.Run("deterministic", func(t *testing.T) {
		h1, err := ComputeBodyHash("Hello")
		if err != nil {
			t.Fatal(err)
		}
		h2, err := ComputeBodyHash("Hello")
		if err != nil {
			t.Fatal(err)
		}
		if h1 != h2 {
			t.Error("same input should produce same hash")
		}
	})

	t.Run("different inputs different hashes", func(t *testing.T) {
		h1, _ := ComputeBodyHash("Hello")
		h2, _ := ComputeBodyHash("World")
		if h1 == h2 {
			t.Error("different inputs should produce different hashes")
		}
	})

	t.Run("hex encoded", func(t *testing.T) {
		h, _ := ComputeBodyHash("test")
		if len(h) != 64 { // SHA-256 = 32 bytes = 64 hex chars
			t.Errorf("expected 64 hex chars, got %d", len(h))
		}
	})
}

func TestIsBodyEncrypted(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"fenc:abc123", true},
		{"fenc:", true},
		{"plain text", false},
		{"enc:other-prefix", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsBodyEncrypted(tt.value)
		if got != tt.want {
			t.Errorf("IsBodyEncrypted(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
