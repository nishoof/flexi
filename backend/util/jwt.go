package util

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

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

func GetUserIdFromJWT(tokenString string) (int64, error) {
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

func VerifyJWT(tokenString string) bool {
	token, err := jwt.Parse(tokenString, keyFunc)
	return err == nil && token.Valid
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	return GetByteKey()
}
