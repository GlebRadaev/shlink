package utils

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

const (
	NameCookieUserID = "user_id"
	SecretKey        = "secret_key"
	TokenExp         = time.Hour * 3
)

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

func CreateCookie(name, value string) *http.Cookie {
	return &http.Cookie{
		Name:  name,
		Value: value,
	}
}

func SetUserIDInCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	userID := GenerateUUID()
	tokenString, _ := GenerateJWT(userID)
	cookie := CreateCookie(NameCookieUserID, tokenString)
	http.SetCookie(w, cookie)
	return userID, nil
}

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

func GetOrSetUserIDFromCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	if userID, ok := GetUserIDFromCookie(r); ok {
		return userID, nil
	}
	return SetUserIDInCookie(w, r)
}
