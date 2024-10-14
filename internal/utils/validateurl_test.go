package utils

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	type args struct {
		urlStr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid http URL",
			args:    args{urlStr: "http://example.com"},
			wantErr: false,
		},
		{
			name:    "valid https URL",
			args:    args{urlStr: "https://example.com"},
			wantErr: false,
		},
		{
			name:    "invalid URL scheme",
			args:    args{urlStr: "ftp://example.com"},
			wantErr: true,
			errMsg:  "invalid URL scheme",
		},
		{
			name:    "invalid URL format",
			args:    args{urlStr: "://invalid-url"},
			wantErr: true,
			errMsg:  "invalid URL format",
		},
		// {
		// 	name:    "invalid domain name",
		// 	args:    args{urlStr: "http://invalid.domain"},
		// 	wantErr: true,
		// 	errMsg:  "invalid domain name",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateURL(tt.args.urlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
