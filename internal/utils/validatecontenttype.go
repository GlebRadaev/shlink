package utils

import (
	"errors"
	"net/http"
	"strings"
)

func ValidateContentType(w http.ResponseWriter, r *http.Request, allowedTypes ...string) error {
	contentType := r.Header.Get("Content-Type")
	for _, allowedType := range allowedTypes {
		if strings.HasPrefix(contentType, allowedType) {
			return nil
		}
	}
	return errors.New("Invalid content type")
}
