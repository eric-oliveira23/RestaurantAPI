package hash_generator

import (
	"crypto/rand"
	"encoding/hex"
)

func HashGenerator(length int) (string, error) {

	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomHash := hex.EncodeToString(randomBytes)

	return randomHash, nil
}
