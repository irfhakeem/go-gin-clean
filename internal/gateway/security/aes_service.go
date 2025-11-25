package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"go-gin-clean/pkg/config"
)

type AESService struct {
	cfg *config.AESConfig
}

func NewAESService(cfg *config.AESConfig) *AESService {
	return &AESService{cfg: cfg}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("ciphertext is empty")
	}
	unpadding := int(data[length-1])
	if unpadding > aes.BlockSize || unpadding == 0 || unpadding > length {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := length - unpadding; i < length; i++ {
		if data[i] != byte(unpadding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:(length - unpadding)], nil
}

func (a *AESService) EncryptInternal(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher([]byte(a.cfg.Key))
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, []byte(a.cfg.IV))
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *AESService) DecryptInternal(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(a.cfg.Key))
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(a.cfg.IV))
	mode.CryptBlocks(ciphertextBytes, ciphertextBytes)

	plaintext, err := pkcs7Unpad(ciphertextBytes)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (a *AESService) EncryptURLSafe(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher([]byte(a.cfg.Key))
	if err != nil {
		return "", err
	}

	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, []byte(a.cfg.IV))
	mode.CryptBlocks(ciphertext, padded)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (a *AESService) DecryptURLSafe(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(a.cfg.Key))
	if err != nil {
		return "", err
	}

	if len(ciphertextBytes)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(a.cfg.IV))
	mode.CryptBlocks(ciphertextBytes, ciphertextBytes)

	plaintext, err := pkcs7Unpad(ciphertextBytes)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
