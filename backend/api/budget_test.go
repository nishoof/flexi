package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBudgetHandler(t *testing.T) {
	rr := sendBudgetRequestAuthed(http.MethodGet, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET budget: expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestBudgetHandlerAuth(t *testing.T) {
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
			rr := sendBudgetRequest(st.method, nil, cookie)

			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestBudgetHandlerMethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodPatch, http.MethodDelete, http.MethodPost}
	for _, method := range methods {
		t.Run(method, func(tt *testing.T) {
			rr := sendBudgetRequestAuthed(method, nil)
			assertStatusAndBody(tt, http.StatusMethodNotAllowed, rr.Code, rr.Body)
		})
	}
}

func TestBudgetHandlerPUT(t *testing.T) {
	body := map[string]interface{}{
		"holidays": []string{},
	}
	bodyBytes, _ := json.Marshal(body)

	rr := sendBudgetRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)
}

func sendBudgetRequestAuthed(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendBudgetRequest(method, body, testJWTTokenCookie)
}

func sendBudgetRequest(method string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendRequest(method, "/api/budget", body, auth, BudgetHandler)
}
