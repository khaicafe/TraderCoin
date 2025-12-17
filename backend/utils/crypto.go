package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// JWTSecret - Shared JWT secret key for all controllers
// TODO: Move this to environment variable or config in production
var JWTSecret = []byte("your-super-secret-jwt-key-change-this-in-production")

// EncryptionKey - Key for encrypting sensitive data (API keys, secrets)
// This will be set from config during initialization
// Must be 32 bytes for AES-256
var EncryptionKey []byte

// InitEncryptionKey sets the encryption key from config
func InitEncryptionKey(key string) {
	EncryptionKey = []byte(key)
	// Ensure key is exactly 32 bytes for AES-256
	if len(EncryptionKey) < 32 {
		// Pad with zeros if too short
		padded := make([]byte, 32)
		copy(padded, EncryptionKey)
		EncryptionKey = padded
	} else if len(EncryptionKey) > 32 {
		// Truncate if too long
		EncryptionKey = EncryptionKey[:32]
	}
}

// EncryptString encrypts a string using AES-256-GCM
func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(EncryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string encrypted with EncryptString
func DecryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(EncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
