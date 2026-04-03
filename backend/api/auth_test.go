package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHandler(t *testing.T) {
	subTests := []struct {
		name         string
		body         interface{}
		expectedCode int
	}{
		{"valid credential with mock", map[string]string{"credential": "mock-google-jwt"}, http.StatusUnauthorized},
		{"missing credential", map[string]interface{}{}, http.StatusBadRequest},
		{"empty credential", map[string]string{"credential": ""}, http.StatusBadRequest},
		{"invalid JSON", "not-json", http.StatusBadRequest},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			var bodyBytes []byte
			switch v := st.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(st.body)
				if err != nil {
					tt.Fatalf("Failed to marshal body: %v", err)
				}
			}

			rr := sendAuthRequest(http.MethodPost, bytes.NewReader(bodyBytes))
			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestAuthHandlerMethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPatch, http.MethodDelete, http.MethodPut}
	for _, method := range methods {
		t.Run(method, func(tt *testing.T) {
			rr := sendAuthRequest(method, nil)
			assertStatusAndBody(tt, http.StatusMethodNotAllowed, rr.Code, rr.Body)
		})
	}
}

func sendAuthRequest(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendRequest(method, "/api/auth", body, nil, AuthHandler)
}
