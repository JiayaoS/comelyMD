package storage

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const pwdCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateID returns a cryptographically random Base62 string with the requested length.
func GenerateID(length int) (string, error) {
	b := make([]byte, length)
	charLen := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

// GeneratePassword returns a cryptographically random password with the requested length.
func GeneratePassword(length int) (string, error) {
	b := make([]byte, length)
	charLen := big.NewInt(int64(len(pwdCharset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			return "", err
		}
		b[i] = pwdCharset[n.Int64()]
	}
	return string(b), nil
}
