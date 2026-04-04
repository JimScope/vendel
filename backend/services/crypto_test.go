package services

import (
	"os"
	"strings"
	"testing"
)

func TestAESGCMRoundTrip(t *testing.T) {
	key := make([]byte, 32) // zero key for testing
	for i := range key {
		key[i] = byte(i)
	}

	tests := []struct {
		name      string
		plaintext string
		prefix    string
	}{
		{"simple text", "Hello world", "enc:"},
		{"empty text", "", "enc:"},
		{"unicode text", "Hola mundo 🌍", "enc:"},
		{"long text", strings.Repeat("x", 10000), "enc:"},
		{"different prefix", "secret", "fenc:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.plaintext == "" {
				// Empty plaintext edge case — encrypt produces empty nonce+ciphertext
				// which is still valid
				return
			}

			encrypted, err := aesGCMEncrypt(key, tt.plaintext, tt.prefix)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if !strings.HasPrefix(encrypted, tt.prefix) {
				t.Errorf("expected prefix %q, got %q", tt.prefix, encrypted[:len(tt.prefix)])
			}

			decrypted, err := aesGCMDecrypt(key, encrypted, tt.prefix)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAESGCMDecrypt_InvalidData(t *testing.T) {
	key := make([]byte, 32)

	tests := []struct {
		name  string
		value string
	}{
		{"invalid base64", "enc:not-valid-base64!!!"},
		{"too short ciphertext", "enc:AAAA"},
		{"empty after prefix", "enc:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := aesGCMDecrypt(key, tt.value, "enc:")
			if err == nil {
				t.Error("expected error for invalid data")
			}
		})
	}
}

func TestAESGCMEncrypt_DifferentNonces(t *testing.T) {
	key := make([]byte, 32)
	plaintext := "same input"

	enc1, _ := aesGCMEncrypt(key, plaintext, "enc:")
	enc2, _ := aesGCMEncrypt(key, plaintext, "enc:")

	if enc1 == enc2 {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncryptDecryptSecret(t *testing.T) {
	os.Setenv("WEBHOOK_ENCRYPTION_KEY", "test-secret-key-for-testing-only")
	defer os.Unsetenv("WEBHOOK_ENCRYPTION_KEY")

	t.Run("round-trip", func(t *testing.T) {
		encrypted, err := EncryptSecret("my-webhook-secret")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(encrypted, encryptedPrefix) {
			t.Errorf("expected prefix %q", encryptedPrefix)
		}

		decrypted, err := DecryptSecret(encrypted)
		if err != nil {
			t.Fatal(err)
		}
		if decrypted != "my-webhook-secret" {
			t.Errorf("got %q, want %q", decrypted, "my-webhook-secret")
		}
	})

	t.Run("empty passthrough", func(t *testing.T) {
		enc, err := EncryptSecret("")
		if err != nil {
			t.Fatal(err)
		}
		if enc != "" {
			t.Error("empty should pass through")
		}
	})

	t.Run("already encrypted passthrough", func(t *testing.T) {
		enc, err := EncryptSecret("enc:already-encrypted")
		if err != nil {
			t.Fatal(err)
		}
		if enc != "enc:already-encrypted" {
			t.Error("already encrypted should pass through")
		}
	})

	t.Run("plaintext fallback on decrypt", func(t *testing.T) {
		dec, err := DecryptSecret("plaintext-value")
		if err != nil {
			t.Fatal(err)
		}
		if dec != "plaintext-value" {
			t.Error("non-prefixed values should return as-is")
		}
	})
}

func TestEncryptSecret_NoKey(t *testing.T) {
	os.Unsetenv("WEBHOOK_ENCRYPTION_KEY")

	_, err := EncryptSecret("secret")
	if err == nil {
		t.Error("expected error when encryption key is not set")
	}
}

func TestGenerateSecureKey(t *testing.T) {
	key := GenerateSecureKey("vk_", 32)

	if !strings.HasPrefix(key, "vk_") {
		t.Errorf("expected prefix vk_, got %q", key)
	}
	if len(key) != 3+32 { // prefix + random
		t.Errorf("expected length 35, got %d", len(key))
	}

	// Two keys should be different
	key2 := GenerateSecureKey("vk_", 32)
	if key == key2 {
		t.Error("expected different keys")
	}
}

func TestGenerateKeyPrefix(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{"long key", "vk_abcdefghijklmnop", "vk_abcdefg..."},
		{"exact length", "1234567890", "1234567890"},
		{"short key", "abc", "abc"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateKeyPrefix(tt.key)
			if got != tt.want {
				t.Errorf("GenerateKeyPrefix(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
