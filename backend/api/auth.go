package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nishoof/flexi/backend/util"
	"google.golang.org/api/idtoken"
)

const jwtExpirationSeconds = 24 * 60 * 60 // 24 hours

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	isOptionsRequest := util.HandleCORS(w, r)
	if isOptionsRequest {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	credential, err := extractCredentialFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if credential == "" {
		http.Error(w, "Credential is required", http.StatusBadRequest)
		return
	}

	// Verify the Google JWT token
	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	payload, err := idtoken.Validate(context.Background(), credential, clientID)
	if err != nil {
		http.Error(w, "Invalid Google credential", http.StatusUnauthorized)
		return
	}

	// Extract user information from the payload
	email := payload.Claims["email"].(string)
	if email == "" {
		http.Error(w, "Email not found in Google token", http.StatusUnauthorized)
		return
	}

	// Generate our own JWT
	token, err := generateJWT(email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		fmt.Println("Error generating JWT:", err)
		return
	}

	// Set JWT as httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   jwtExpirationSeconds,
		HttpOnly: true,                  // Not accessible via JavaScript
		Secure:   true,                  // Only send over HTTPS
		SameSite: http.SameSiteNoneMode, // CSRF protection
	})

	fmt.Println("User authenticated:", email)
	fmt.Println("Generated JWT:", token)
}

func extractCredentialFromRequest(r *http.Request) (string, error) {
	var contents struct {
		Credential string `json:"credential"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&contents)
	if err != nil {
		return "", err
	}

	return contents.Credential, nil
}

func generateJWT(email string) (string, error) {
	const tokenExpiration = jwtExpirationSeconds * time.Second

	key := os.Getenv("JWT_KEY")
	if key == "" {
		fmt.Println("JWT_KEY environment variable is not set")
		return "", jwt.ErrInvalidKey
	}
	byteKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Println("Error decoding JWT key:", err)
		return "", err
	}

	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(tokenExpiration).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(byteKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
