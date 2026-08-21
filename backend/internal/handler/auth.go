package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nishoof/flexi/backend/internal/apierr"
	"github.com/nishoof/flexi/backend/internal/database"
	"github.com/nishoof/flexi/backend/internal/util"
	"google.golang.org/api/idtoken"
)

const jwtExpiration = 24 * time.Hour
const noUserId = -1

func RegisterAuth(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth", Handle(AuthHandler))
}

func AuthHandler(w http.ResponseWriter, r *http.Request) *apierr.Error {
	credential, err := extractCredentialFromRequest(r)
	if err != nil {
		return apierr.BadRequest("Invalid request body")
	}
	if credential == "" {
		return apierr.BadRequest("Credential is required")
	}

	// Verify the Google JWT token
	ctx := r.Context()
	googleClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	payload, err := idtoken.Validate(ctx, credential, googleClientID)
	if err != nil {
		return apierr.Unauthorized("Invalid Google credential")
	}

	// Extract user information from the payload
	email := payload.Claims["email"].(string)
	if email == "" {
		return apierr.Unauthorized("Email not found in Google token")
	}

	// Check if user exists in the database, create if not
	userId, err := getOrCreateUser(ctx, email)
	if err != nil {
		return apierr.Internal(err)
	}

	// Generate our own JWT
	token, err := generateJWT(userId)
	if err != nil {
		return apierr.Internal(err)
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
	return nil
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
