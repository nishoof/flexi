package util

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper functions to create test JWTs with specific claims (e.g. userId, expiration)
func createTestJWTWithClaims(secret string, claims jwt.MapClaims) string {
	key, _ := base64.StdEncoding.DecodeString(secret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(key)
	return tokenString
}

func createTestJWT(userId int64, secret string) string {
	return createTestJWTWithClaims(secret, jwt.MapClaims{
		"userId": float64(userId),
		"exp":    time.Now().Add(time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	})
}

func createExpiredTestJWT(userId int64, secret string) string {
	return createTestJWTWithClaims(secret, jwt.MapClaims{
		"userId": float64(userId),
		"exp":    time.Now().Add(-time.Hour).Unix(),
		"iat":    time.Now().Add(-2 * time.Hour).Unix(),
	})
}

func createJWTWithoutUserId(secret string) string {
	return createTestJWTWithClaims(secret, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
}

func TestVerifyJWT(t *testing.T) {
	secret := "dGVzdGtl" // "testke" in base64
	t.Setenv("JWT_KEY", secret)

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid JWT",
			token: createTestJWT(1, secret),
			want:  true,
		},
		{
			name:  "malformed JWT",
			token: "not-a-valid-jwt-token",
			want:  false,
		},
		{
			name:  "empty token",
			token: "",
			want:  false,
		},
		{
			name:  "invalid signature",
			token: createTestJWT(1, "c2VjcmV0"), // "secret" different key
			want:  false,
		},
		{
			name:  "expired JWT",
			token: createExpiredTestJWT(1, secret),
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := verifyJWT(test.token)
			if got != test.want {
				t.Errorf("verifyJWT(%q) = %v, want %v", test.token, got, test.want)
			}
		})
	}
}

func TestGetUserIdFromJWT(t *testing.T) {
	secret := "dGVzdGtl"
	t.Setenv("JWT_KEY", secret)

	tests := []struct {
		name       string
		token      string
		wantUserId int64
		wantErr    bool
	}{
		{
			name:       "valid token with userId",
			token:      createTestJWT(123, secret),
			wantUserId: 123,
			wantErr:    false,
		},
		{
			name:       "token without userId claim",
			token:      createJWTWithoutUserId(secret),
			wantUserId: -1,
			wantErr:    true,
		},
		{
			name:       "invalid token",
			token:      "invalid.token.here",
			wantUserId: -1,
			wantErr:    true,
		},
		{
			name:       "expired token",
			token:      createExpiredTestJWT(123, secret),
			wantUserId: -1,
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userId, err := getUserIdFromJWT(test.token)
			if test.wantErr && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !test.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			} else if !test.wantErr && userId != test.wantUserId {
				t.Errorf("Expected userId %d but got %d", test.wantUserId, userId)
			}
		})
	}
}

func TestAuthenticateUser(t *testing.T) {
	secret := "dGVzdGtl"
	t.Setenv("JWT_KEY", secret)

	tests := []struct {
		name        string
		cookieValue string
		wantUserId  int64
		wantErr     bool
	}{
		{
			name:        "valid token",
			cookieValue: createTestJWT(42, secret),
			wantUserId:  42,
			wantErr:     false,
		},
		{
			name:        "missing cookie",
			cookieValue: "",
			wantUserId:  -1,
			wantErr:     true,
		},
		{
			name:        "invalid token",
			cookieValue: "invalid.token",
			wantUserId:  -1,
			wantErr:     true,
		},
		{
			name:        "expired token",
			cookieValue: createExpiredTestJWT(42, secret),
			wantUserId:  -1,
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &http.Request{
				Header: make(http.Header),
			}
			if test.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: test.cookieValue})
			}

			userId, err := AuthenticateUser(req)
			if test.wantErr && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !test.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			} else if !test.wantErr && userId != test.wantUserId {
				t.Errorf("Expected userId %d but got %d", test.wantUserId, userId)
			}
		})
	}
}
