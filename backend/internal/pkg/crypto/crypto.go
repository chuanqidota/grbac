package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt work factor used for password hashing.
const DefaultCost = 12

// HashPassword generates a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a bcrypt-hashed password with a plaintext candidate.
// Returns nil on match.
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateHMAC computes an HMAC-SHA256 signature of the given payload using
// the provided secret key, returned as a hex-encoded string.
// Suitable for webhook signature generation.
func GenerateHMAC(payload []byte, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC checks whether the given hex-encoded signature matches the
// HMAC-SHA256 of the payload under the secret key.
func VerifyHMAC(payload []byte, secret []byte, signature string) bool {
	expected := GenerateHMAC(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// Encrypt encrypts plaintext using AES-GCM with the given key (must be 16, 24, or 32 bytes).
// Returns hex-encoded nonce + ciphertext.
func Encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts hex-encoded nonce+ciphertext using AES-GCM with the given key.
func Decrypt(encoded string, key []byte) ([]byte, error) {
	data, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("aesGCM.Open: %w", err)
	}

	return plaintext, nil
}
