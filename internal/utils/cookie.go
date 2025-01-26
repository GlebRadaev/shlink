// Package utils provides utility functions for handling JWT tokens, cookies,
// and managing user authentication-related actions in a web application.
package utils

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure, including standard claims from
// the JWT specification and custom user data like UserID.
type Claims struct {
	jwt.RegisteredClaims        // Standard JWT claims (issued at, expiration, etc.)
	UserID               string // Custom user identifier for the JWT.
}

// Constants used for handling JWT tokens and cookies in the application.
const (
	NameCookieUserID = "user_id"     // Name of the cookie where the user ID is stored.
	SecretKey        = "secret_key"  // Secret key used to sign JWT tokens.
	TokenExp         = time.Hour * 3 // Expiration time for the JWT token.
)

// GenerateJWT creates a JWT token containing the provided user ID with a 3-hour expiration time.
// The token is signed with a secret key and returned as a string.
func GenerateJWT(userID string) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}

// ParseJWT parses and validates a JWT token string. It uses the secret key to verify the signature
// and returns an error if the token is invalid or expired.
func ParseJWT(tokenString string, claims *Claims) error {
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("expected signing method: " + jwt.SigningMethodHS256.Alg())
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		return err
	}
	if !parsedToken.Valid {
		return errors.New("invalid token")
	}
	return nil
}

// CreateCookie creates and returns an HTTP cookie with the specified name and value.
func CreateCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:  name,
		Value: value,
	}
}

// SetUserIDInCookie generates a new user ID, creates a JWT with that ID, and sets the token as a cookie
// in the response. It returns the generated user ID.
func SetUserIDInCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	userID := GenerateUUID()
	tokenString, _ := GenerateJWT(userID)
	cookie := CreateCookie(NameCookieUserID, tokenString)
	http.SetCookie(w, cookie)
	return userID, nil
}

// GetUserIDFromCookie retrieves the user ID from the cookie in the request.
func GetUserIDFromCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(NameCookieUserID)
	if err != nil {
		return "", false
	}
	claims := &Claims{}
	if err := ParseJWT(cookie.Value, claims); err != nil {
		return "", false
	}
	return claims.UserID, true
}

// GetOrSetUserIDFromCookie checks if a valid user ID is present in the cookie. If a valid ID is found,
// it returns the user ID.
func GetOrSetUserIDFromCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	if userID, ok := GetUserIDFromCookie(r); ok {
		return userID, nil
	}
	return SetUserIDInCookie(w, r)
}
