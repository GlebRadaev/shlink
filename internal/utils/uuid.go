package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateUUID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
