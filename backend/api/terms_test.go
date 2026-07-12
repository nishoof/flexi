package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nishoof/flexi/backend/database"
)

func TestTermsHandler(t *testing.T) {
	rr := sendTermRequestAuthed(http.MethodGet, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET term: expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestTermsHandlerAuth(t *testing.T) {
	subTests := []struct {
		name         string
		method       string
		authToken    string
		expectedCode int
	}{
		{"GET without auth token", http.MethodGet, "", http.StatusUnauthorized},
		{"PUT without auth token", http.MethodPut, "", http.StatusUnauthorized},
		{"GET with invalid auth token", http.MethodGet, "invalid-token", http.StatusUnauthorized},
		{"PUT with invalid auth token", http.MethodPut, "invalid-token", http.StatusUnauthorized},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			var cookie *http.Cookie
			if st.authToken != "" {
				cookie = &http.Cookie{Name: "auth_token", Value: st.authToken}
			}
			rr := sendTermRequest(st.method, nil, cookie)

			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestTermsHandlerMethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodPatch, http.MethodDelete, http.MethodPost}
	for _, method := range methods {
		t.Run(method, func(tt *testing.T) {
			rr := sendTermRequestAuthed(method, nil)
			assertStatusAndBody(tt, http.StatusMethodNotAllowed, rr.Code, rr.Body)
		})
	}
}

func TestTermsHandlerPUT(t *testing.T) {
	registerTermCleanup(t, testUserId)

	// No days off campus

	body := map[string]interface{}{
		"name":     "Spring 2026",
		"days_off": []string{},
	}
	bodyBytes, _ := json.Marshal(body)

	rr := sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	// One day off campus

	body = map[string]interface{}{
		"name":     "Spring 2026",
		"days_off": []string{"2026-07-31"},
	}
	bodyBytes, _ = json.Marshal(body)

	rr = sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	// Multiple days off campus

	body = map[string]interface{}{
		"name":     "Spring 2026",
		"days_off": []string{"2026-07-31", "2026-04-06"},
	}
	bodyBytes, _ = json.Marshal(body)

	rr = sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)
}

func sendTermRequestAuthed(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendTermRequest(method, body, testJWTTokenCookie)
}

func sendTermRequest(method string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendRequest(method, "/api/terms", body, auth, TermsHandler)
}

func registerTermCleanup(t testing.TB, userId int64) {
	queries, err := database.Queries(context.Background())
	if err != nil {
		t.Fatalf("Failed to get database queries: %v", err)
	}

	t.Cleanup(func() {
		if err := queries.DeleteActiveTermByUser(context.Background(), userId); err != nil {
			t.Fatalf("Failed to clean up term: %v", err)
		}
	})
}
