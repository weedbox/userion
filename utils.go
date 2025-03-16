package userion

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateSalt creates a random salt for password hashing
func GenerateSalt() (string, error) {
	// Generate a random 16-byte salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Convert to hex string
	return hex.EncodeToString(salt), nil
}

// HashPassword combines a password with a salt and hashes it
func HashPassword(password, salt string) string {
	// Combine password and salt
	combined := password + salt

	// Hash using SHA-256
	hash := sha256.New()
	hash.Write([]byte(combined))

	// Return hex encoded hash
	return hex.EncodeToString(hash.Sum(nil))
}
