// Package utils provides utility functions for generating random strings
// and validating string-based IDs with specific character sets and lengths.
package utils

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generate generates a random string of the specified length, using
// characters from the charset constant (letters and digits).
func Generate(length int) string {
	const charLen = len(charset)
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(charLen)]
	}
	return string(result)
}

// IsValidID checks if the given ID string has the exact specified length
// and contains only characters from the allowed charset (letters and digits).
func IsValidID(id string, length int) bool {
	if len(id) != length {
		return false
	}
	for _, char := range id {
		if !strings.ContainsRune(charset, char) {
			return false
		}
	}
	return true
}
