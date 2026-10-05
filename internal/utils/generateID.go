package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateId() (string, error) {
	b := make([]byte, 23)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}