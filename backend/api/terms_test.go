package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
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
	// No days off campus

	body := map[string]interface{}{
		"name":     "Spring 2026",
		"end_date": "2026-05-23",
		"days_off": []string{},
	}
	bodyBytes, _ := json.Marshal(body)

	rr := sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	// One day off campus

	body = map[string]interface{}{
		"name":     "Spring 2026",
		"end_date": "2026-05-23",
		"days_off": []string{"2026-07-31"},
	}
	bodyBytes, _ = json.Marshal(body)

	rr = sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	// Multiple days off campus

	body = map[string]interface{}{
		"name":     "Spring 2026",
		"end_date": "2026-05-23",
		"days_off": []string{"2026-07-31", "2026-04-06"},
	}
	bodyBytes, _ = json.Marshal(body)

	rr = sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)
}

func TestTermsHandlerPUTValidation(t *testing.T) {
	subTests := []struct {
		name         string
		body         interface{}
		expectedCode int
	}{
		{"missing end_date", map[string]interface{}{
			"name":     "Spring 2026",
			"days_off": []string{},
		}, http.StatusBadRequest},
		{"invalid end_date", map[string]interface{}{
			"name":     "Spring 2026",
			"end_date": "not-a-date",
			"days_off": []string{},
		}, http.StatusBadRequest},
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

			rr := sendTermRequestAuthed(http.MethodPut, bytes.NewReader(bodyBytes))
			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func sendTermRequestAuthed(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendTermRequest(method, body, testJWTTokenCookie)
}

func sendTermRequest(method string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendRequest(method, "/api/terms", body, auth, TermsHandler)
}
