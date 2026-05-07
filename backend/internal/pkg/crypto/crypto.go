package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

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
