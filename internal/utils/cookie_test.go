package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		wantErr bool
	}{
		{"Valid userID", "test_user_id", false},
		{"Empty userID", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GenerateJWT(tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		setupToken func() string
		wantErr    bool
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := GenerateJWT("test_user_id")
				return token
			},
			wantErr: false,
		},
		{
			name:       "Invalid token",
			setupToken: func() string { return "invalid.token.string" },
			wantErr:    true,
		},
		{
			name: "Invalid signing method",
			setupToken: func() string {
				return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF91c2VyX2lkIn0.invalid_signature"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.setupToken()
			claims := &Claims{}
			err := ParseJWT(token, claims)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "test_user_id", claims.UserID)
			}
		})
	}
}

func TestCreateCookie(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"Valid cookie", "test_cookie", "test_value"},
		{"Empty key and value", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookie := CreateCookie(tc.key, tc.value)
			assert.Equal(t, tc.key, cookie.Name)
			assert.Equal(t, tc.value, cookie.Value)
		})
	}
}

func TestSetUserIDInCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	userID, err := SetUserIDInCookie(rec, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, userID)

	resp := rec.Result()
	defer resp.Body.Close()
	assert.NotEmpty(t, resp.Cookies())
	assert.Equal(t, NameCookieUserID, resp.Cookies()[0].Name)
}

func TestGetUserIDFromCookie(t *testing.T) {
	tests := []struct {
		name       string
		setupToken func() string
		wantOK     bool
	}{
		{
			name: "Valid cookie",
			setupToken: func() string {
				token, _ := GenerateJWT("test_user_id")
				return token
			},
			wantOK: true,
		},
		{
			name:       "No cookie",
			setupToken: func() string { return "" },
			wantOK:     false,
		},
		{
			name:       "Invalid token",
			setupToken: func() string { return "invalid_token" },
			wantOK:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.setupToken() != "" {
				req.AddCookie(&http.Cookie{Name: NameCookieUserID, Value: tc.setupToken()})
			}

			userID, ok := GetUserIDFromCookie(req)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.NotEmpty(t, userID)
			} else {
				assert.Empty(t, userID)
			}
		})
	}
}

func TestGetOrSetUserIDFromCookie(t *testing.T) {
	tests := []struct {
		name           string
		setupExisting  bool
		expectSameUser bool
	}{
		{"New user ID", false, false},
		{"Existing user ID", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			var expectedUserID string

			if tc.setupExisting {
				// Установить cookie с JWT
				expectedUserID, _ = SetUserIDInCookie(rec, req)

				resp := rec.Result()
				defer resp.Body.Close()

				cookie := resp.Cookies()[0]
				req.AddCookie(cookie)

				// Декодируем токен, чтобы получить UserID
				claims := &Claims{}
				err := ParseJWT(cookie.Value, claims)
				assert.NoError(t, err)
				assert.Equal(t, expectedUserID, claims.UserID)
			}

			// Получить или установить UserID
			userID, err := GetOrSetUserIDFromCookie(rec, req)
			assert.NoError(t, err)
			assert.NotEmpty(t, userID)

			if tc.setupExisting {
				assert.Equal(t, expectedUserID, userID) // Сравниваем UserID, а не сам JWT
			}
		})
	}
}
