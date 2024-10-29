package utils

import (
	"net/http/httptest"
	"testing"
)

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		allowedTypes []string
		wantErr      bool
	}{
		{
			name:         "valid content type - application/json",
			contentType:  "application/json",
			allowedTypes: []string{"application/json"},
			wantErr:      false,
		},
		{
			name:         "valid content type - text/html",
			contentType:  "text/html",
			allowedTypes: []string{"text/html"},
			wantErr:      false,
		},
		{
			name:         "valid content type - multiple types",
			contentType:  "application/json",
			allowedTypes: []string{"application/json", "text/html"},
			wantErr:      false,
		},
		{
			name:         "invalid content type with wildcard",
			contentType:  "application/json; charset=utf-8",
			allowedTypes: []string{"application/*"},
			wantErr:      true,
		},
		{
			name:         "invalid content type",
			contentType:  "text/plain",
			allowedTypes: []string{"application/json"},
			wantErr:      true,
		},
		{
			name:         "invalid content type with multiple allowed types",
			contentType:  "text/plain",
			allowedTypes: []string{"application/json", "text/html"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/", nil)
			r.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()

			err := ValidateContentType(w, r, tt.allowedTypes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
