package utils

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(length int) string {
	const charLen = len(charset)
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(charLen)]
	}
	return string(result)
}

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
