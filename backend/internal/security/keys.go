package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GenerateRSAPrivateKeyPEM generates a new 2048-bit RSA private key in PEM format.
func GenerateRSAPrivateKeyPEM() (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}
	return string(pem.EncodeToMemory(privBlock)), nil
}
