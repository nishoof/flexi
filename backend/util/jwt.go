package util

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// AuthenticateUser checks the "auth_token" cookie for a valid JWT and
// returns the user ID if valid.
func AuthenticateUser(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return -1, fmt.Errorf("Unauthorized: Missing token %w", err)
	}

	jwt := cookie.Value
	valid := verifyJWT(jwt)
	if !valid {
		return -1, fmt.Errorf("Unauthorized: Invalid token")
	}

	userId, err := getUserIdFromJWT(jwt)
	if err != nil {
		return -1, fmt.Errorf("Unauthorized: %w", err)
	}

	return userId, nil
}

// GetByteKey retrieves the JWT signing key from the environment
// variable and decodes it from base64.
func GetByteKey() ([]byte, error) {
	key := os.Getenv("JWT_KEY")
	if key == "" {
		fmt.Println("JWT_KEY environment variable is not set")
		return nil, jwt.ErrInvalidKey
	}

	byteKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Println("Error decoding JWT key:", err)
		return nil, err
	}

	return byteKey, nil
}

func getUserIdFromJWT(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return -1, err
	}
	if !token.Valid {
		return -1, fmt.Errorf("Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, jwt.ErrTokenInvalidClaims
	}

	userIdFloat, ok := claims["userId"].(float64)
	if !ok {
		return -1, jwt.ErrTokenInvalidClaims
	}

	return int64(userIdFloat), nil
}

func verifyJWT(tokenString string) bool {
	token, err := jwt.Parse(tokenString, keyFunc)
	return err == nil && token.Valid
}

// Provides the signing key. Used by jwt.Parse
// https://pkg.go.dev/github.com/golang-jwt/jwt/v5@v5.3.0#Keyfunc
func keyFunc(token *jwt.Token) (interface{}, error) {
	return GetByteKey()
}
