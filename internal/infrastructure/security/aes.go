package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"
)

type aesEncryptor struct {
	key [32]byte
}

func NewAESEncryptor(cfg *config.AESConfig) port.Encryptor {
	// Standardize to 32-byte AES-256 key using SHA-256
	key := sha256.Sum256([]byte(cfg.Key))
	return &aesEncryptor{key: key}
}

func (a *aesEncryptor) EncryptInternal(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	sealed, err := a.encryptGCM([]byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (a *aesEncryptor) DecryptInternal(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plain, err := a.decryptGCM(data)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

func (a *aesEncryptor) EncryptURLSafe(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	sealed, err := a.encryptGCM([]byte(plaintext))
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(sealed), nil
}

func (a *aesEncryptor) DecryptURLSafe(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	plain, err := a.decryptGCM(data)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}

func (a *aesEncryptor) encryptGCM(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key[:])
	if err != nil {
		return nil, fmt.Errorf("aes new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends ciphertext and authentication tag to nonce: [nonce (12b)] + [ciphertext] + [tag (16b)]
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (a *aesEncryptor) decryptGCM(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key[:])
	if err != nil {
		return nil, fmt.Errorf("aes new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher new gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("gcm open authentication failed: %w", err)
	}

	return plaintext, nil
}
