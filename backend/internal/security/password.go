package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Default parameters for Argon2id
var defaultParams = Params{
	Memory:      32 * 1024, // 32MB
	Iterations:  1,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// HashPassword hashes a plain text password using Argon2id.
func HashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, defaultParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultParams.Iterations,
		defaultParams.Memory,
		defaultParams.Parallelism,
		defaultParams.KeyLength,
	)

	// Encode to standard modular crypt format
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultParams.Memory,
		defaultParams.Iterations,
		defaultParams.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// VerifyPassword compares a password against an Argon2id encoded hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	params := Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	params.KeyLength = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}
