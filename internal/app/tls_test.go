package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCertificate(t *testing.T) {
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "test_cert.pem")
	keyPath := filepath.Join(tempDir, "test_key.pem")

	testCases := []struct {
		name        string
		certPath    string
		keyPath     string
		prepare     func()
		expectedErr bool
	}{
		{
			name:        "Successful certificate generation",
			certPath:    certPath,
			keyPath:     keyPath,
			prepare:     func() {},
			expectedErr: false,
		},
		{
			name:     "Certificate already exists",
			certPath: certPath,
			keyPath:  keyPath,
			prepare: func() {
				_ = GenerateCertificate(certPath, keyPath)
			},
			expectedErr: false,
		},
		// {
		// 	name:        "Invalid directory path",
		// 	certPath:    "/invalid_path/test_cert.pem",
		// 	keyPath:     keyPath,
		// 	prepare:     func() {},
		// 	expectedErr: true,
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepare()

			err := GenerateCertificate(tc.certPath, tc.keyPath)

			if (err != nil) != tc.expectedErr {
				t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
			}

			if !tc.expectedErr {
				if _, err := os.Stat(tc.certPath); os.IsNotExist(err) {
					t.Errorf("Certificate not found: %s", tc.certPath)
				}
				if _, err := os.Stat(tc.keyPath); os.IsNotExist(err) {
					t.Errorf("Key not found: %s", tc.keyPath)
				}
			}
		})
	}
}
