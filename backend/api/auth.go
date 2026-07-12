package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/util"
	"google.golang.org/api/idtoken"
)

const jwtExpiration = 24 * time.Hour
const noUserId = -1

func AuthHandler(w http.ResponseWriter, r *http.Request) {
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
	ctx := r.Context()
	googleClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	payload, err := idtoken.Validate(ctx, credential, googleClientID)
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

	// Check if user exists in the database, create if not
	userId, err := getOrCreateUser(ctx, email)
	if err != nil {
		fmt.Println("Error in getOrCreateUser:", err)
		http.Error(w, "Failed to get or create user", http.StatusInternalServerError)
		return
	}

	// Generate our own JWT
	token, err := generateJWT(userId)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set JWT as httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(jwtExpiration / time.Second),
		HttpOnly: true, // Not accessible via JavaScript
		Secure:   true, // Only send over HTTPS
		SameSite: http.SameSiteNoneMode,
	})
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

func generateJWT(userId int64) (string, error) {
	const tokenExpiration = jwtExpiration

	byteKey, err := util.GetByteKey()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(tokenExpiration).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(byteKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func getOrCreateUser(ctx context.Context, email string) (int64, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return noUserId, err
	}

	id, err := queries.GetOrCreateUser(ctx, email)
	if err != nil {
		return noUserId, err
	}

	return id, nil
}
