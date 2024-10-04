package utils

import (
	"math/rand"
)

func Generate(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const charLen = len(charset)
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(charLen)]
	}
	return string(result)
}
