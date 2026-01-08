package util

import "net/http"

/* Handles CORS (Cross-Origin Resource Sharing). Sets appropriate headers and responds to OPTIONS (preflight) requests. Returns if the request is an OPTIONS request. If it's an OPTIONS request, then no further processing is needed. */
func HandleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")

	allowed := map[string]bool{
		"http://localhost:5173":         true,
		"https://flexi.nishilanand.com": true,
	}
	if allowed[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	isOptionsRequest := r.Method == http.MethodOptions

	if isOptionsRequest {
		w.WriteHeader(http.StatusOK)
	}

	return isOptionsRequest
}
