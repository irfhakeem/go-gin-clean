package unit_test

import (
	"testing"

	"go-gin-clean/internal/infrastructure/security"
	"go-gin-clean/pkg/config"
)

func TestAESGCM_Encryption(t *testing.T) {
	cfg := &config.AESConfig{
		Key: "test-strong-aes-encryption-key-1234",
	}
	encryptor := security.NewAESEncryptor(cfg)

	t.Run("AC-004: Random Nonce Produces Distinct Ciphertexts", func(t *testing.T) {
		plaintext := "sensitive-user-token-data"

		c1, err1 := encryptor.EncryptInternal(plaintext)
		if err1 != nil {
			t.Fatalf("unexpected error on first encryption: %v", err1)
		}

		c2, err2 := encryptor.EncryptInternal(plaintext)
		if err2 != nil {
			t.Fatalf("unexpected error on second encryption: %v", err2)
		}

		if c1 == c2 {
			t.Fatalf("vulnerability: AES encryption produced identical ciphertext for same plaintext (static IV detected)")
		}

		d1, err := encryptor.DecryptInternal(c1)
		if err != nil || d1 != plaintext {
			t.Fatalf("decryption failed for c1: %v, got %s", err, d1)
		}

		d2, err := encryptor.DecryptInternal(c2)
		if err != nil || d2 != plaintext {
			t.Fatalf("decryption failed for c2: %v, got %s", err, d2)
		}
	})

	t.Run("URLSafe Encryption Roundtrip", func(t *testing.T) {
		plaintext := "user-id-12345_2026-10-07T00:00:00Z"
		ciphertext, err := encryptor.EncryptURLSafe(plaintext)
		if err != nil {
			t.Fatalf("unexpected url safe encryption error: %v", err)
		}

		decrypted, err := encryptor.DecryptURLSafe(ciphertext)
		if err != nil {
			t.Fatalf("unexpected url safe decryption error: %v", err)
		}

		if decrypted != plaintext {
			t.Fatalf("expected %s, got %s", plaintext, decrypted)
		}
	})

	t.Run("Authentication Tag Tampering Fails", func(t *testing.T) {
		plaintext := "important-session-payload"
		ciphertext, err := encryptor.EncryptInternal(plaintext)
		if err != nil {
			t.Fatalf("encryption error: %v", err)
		}

		// Tamper with one byte in the base64 ciphertext
		tamperedBytes := []byte(ciphertext)
		tamperedBytes[len(tamperedBytes)-2] ^= 0xFF
		tampered := string(tamperedBytes)

		_, err = encryptor.DecryptInternal(tampered)
		if err == nil {
			t.Fatalf("expected decryption error on tampered ciphertext, but got nil")
		}
	})
}
