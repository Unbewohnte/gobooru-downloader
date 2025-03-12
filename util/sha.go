package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func SHA256HashFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum(nil)

	return hex.EncodeToString(hash), nil
}
