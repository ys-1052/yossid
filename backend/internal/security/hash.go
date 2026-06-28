package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashWithPepper computes the SHA-256 hash of a string combined with a secret pepper.
func HashWithPepper(value string, pepper string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value + pepper))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateRandomToken generates a URL-safe cryptographically secure random token.
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomOTP generates a 6-digit numeric OTP.
func GenerateRandomOTP() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Derive a 6 digit number from the random bytes
	val := (uint32(bytes[0]) << 24) | (uint32(bytes[1]) << 16) | (uint32(bytes[2]) << 8) | uint32(bytes[3])
	otpNum := (val % 900000) + 100000 // Ensure 6 digits (100000 to 999999)
	return fmt.Sprintf("%06d", otpNum), nil
}
