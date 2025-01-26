// Package utils provides utility functions, including cryptographically secure
// random string generation and other common utilities.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateUUID generates a new UUID (Universally Unique Identifier).
func GenerateUUID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
