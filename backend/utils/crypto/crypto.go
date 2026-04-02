package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/el-bulk/backend/utils/logger"
)

// Encrypt takes a plaintext string and returns a base64-encoded encrypted string.
// It uses AES-GCM for strong, authenticated encryption.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		return "", errors.New("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt takes a base64-encoded encrypted string and returns the plaintext.
func Decrypt(encodedCiphertext string) (string, error) {
	if encodedCiphertext == "" {
		return "", nil
	}

	// If it's not base64, it might be plain text from before encryption was enabled,
	// or it might be corrupted. For migration, we might want to handle this.
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		return "", errors.New("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// DecryptSafe is a helper that logs but doesn't crash on failure, returns the input if decryption fails.
// Useful during transition periods.
func DecryptSafe(val *string) *string {
	if val == nil || *val == "" {
		return val
	}
	decrypted, err := Decrypt(*val)
	if err != nil {
		// Log but return original (maybe it's not encrypted yet)
		logger.Warn("Decryption failed (returning original value): %v", err)
		return val
	}
	return &decrypted
}
