// Package utils provides utility functions, including content type validation,
// cryptographically secure random string generation, and other common utilities.
package utils

import (
	"errors"
	"net/http"
	"strings"
)

// ValidateContentType validates the `Content-Type` header of the incoming HTTP request.
func ValidateContentType(w http.ResponseWriter, r *http.Request, allowedTypes ...string) error {
	contentType := r.Header.Get("Content-Type")
	for _, allowedType := range allowedTypes {
		if strings.HasPrefix(contentType, allowedType) {
			return nil
		}
	}
	return errors.New("invalid content type")
}
