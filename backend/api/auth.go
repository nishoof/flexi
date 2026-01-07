package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/nishoof/flexi/backend/util"
	"google.golang.org/api/idtoken"
)

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
		http.Error(w, "Invalid credential", http.StatusUnauthorized)
		return
	}

	// Extract user information from the payload
	email := payload.Claims["email"].(string)
	if email == "" {
		http.Error(w, "Email not found in token", http.StatusUnauthorized)
		return
	}

	// Respond with the user's email
	response := struct {
		Email string `json:"email"`
	}{
		Email: email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
