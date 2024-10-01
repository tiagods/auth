package cripto

import (
	"crypto/sha256"
	"encoding/hex"
)

func Encode(value string) (string, error) {
	hasher := sha256.New()
	if _, err := hasher.Write([]byte(value)); err != nil {
		return "", err
	}
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash), nil
}
