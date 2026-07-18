package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTermsHandler(t *testing.T) {
	rr := sendTermRequestAuthed(http.MethodGet, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET terms: expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var terms []termResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &terms); err != nil {
		t.Fatalf("GET terms: expected JSON array, got %q: %v", rr.Body.String(), err)
	}
	if len(terms) == 0 {
		t.Fatal("GET terms: expected at least one term")
	}
	activeCount := 0
	for _, term := range terms {
		if term.IsActive {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Fatalf("GET terms: expected exactly one active term, got %d", activeCount)
	}
}

func TestTermsHandlerAuth(t *testing.T) {
	subTests := []struct {
		name         string
		method       string
		path         string
		authToken    string
		expectedCode int
	}{
		{"GET without auth token", http.MethodGet, "/api/terms", "", http.StatusUnauthorized},
		{"PUT without auth token", http.MethodPut, "/api/terms", "", http.StatusUnauthorized},
		{"POST without auth token", http.MethodPost, "/api/terms", "", http.StatusUnauthorized},
		{"GET by id without auth token", http.MethodGet, "/api/terms/1", "", http.StatusUnauthorized},
		{"POST activate without auth token", http.MethodPost, "/api/terms/1/activate", "", http.StatusUnauthorized},
		{"GET with invalid auth token", http.MethodGet, "/api/terms", "invalid-token", http.StatusUnauthorized},
		{"PUT with invalid auth token", http.MethodPut, "/api/terms", "invalid-token", http.StatusUnauthorized},
		{"POST with invalid auth token", http.MethodPost, "/api/terms", "invalid-token", http.StatusUnauthorized},
		{"GET by id with invalid auth token", http.MethodGet, "/api/terms/1", "invalid-token", http.StatusUnauthorized},
		{"POST activate with invalid auth token", http.MethodPost, "/api/terms/1/activate", "invalid-token", http.StatusUnauthorized},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			var cookie *http.Cookie
			if st.authToken != "" {
				cookie = &http.Cookie{Name: "auth_token", Value: st.authToken}
			}
			rr := sendTermRequestPath(st.method, st.path, nil, cookie)

			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestTermsHandlerMethodNotAllowed(t *testing.T) {
	subTests := []struct {
		name   string
		method string
		path   string
	}{
		{"PATCH collection", http.MethodPatch, "/api/terms"},
		{"DELETE collection", http.MethodDelete, "/api/terms"},
		{"PUT by id", http.MethodPut, "/api/terms/1"},
		{"POST by id", http.MethodPost, "/api/terms/1"},
		{"DELETE by id", http.MethodDelete, "/api/terms/1"},
		{"GET activate", http.MethodGet, "/api/terms/1/activate"},
		{"PUT activate", http.MethodPut, "/api/terms/1/activate"},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			rr := sendTermRequestAuthedPath(st.method, st.path, nil)
			assertStatusAndBody(tt, http.StatusMethodNotAllowed, rr.Code, rr.Body)
		})
	}
}

func TestTermsHandlerNotFoundPath(t *testing.T) {
	subTests := []struct {
		name   string
		method string
		path   string
	}{
		{"non-numeric id", http.MethodGet, "/api/terms/abc"},
		{"zero id", http.MethodGet, "/api/terms/0"},
		{"negative id", http.MethodGet, "/api/terms/-1"},
		{"unknown suffix", http.MethodGet, "/api/terms/1/entries"},
		{"extra segment", http.MethodPost, "/api/terms/1/activate/extra"},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			rr := sendTermRequestAuthedPath(st.method, st.path, nil)
			assertStatusAndBody(tt, http.StatusNotFound, rr.Code, rr.Body)
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

func TestTermsHandlerPOSTCreate(t *testing.T) {
	body := map[string]interface{}{
		"name":     "Fall 2026",
		"end_date": "2026-12-15",
		"days_off": []string{"2026-11-26"},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	rr := sendTermRequestAuthed(http.MethodPost, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	var created termResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("POST create: expected term JSON, got %q: %v", rr.Body.String(), err)
	}
	if created.ID == 0 {
		t.Fatal("POST create: expected non-zero id")
	}
	if created.Name != "Fall 2026" {
		t.Fatalf("POST create: expected name Fall 2026, got %q", created.Name)
	}
	if created.EndDate == nil || created.EndDate.String() != "2026-12-15" {
		t.Fatalf("POST create: expected end_date 2026-12-15, got %v", created.EndDate)
	}
	if created.IsActive {
		t.Fatal("POST create: new term should be inactive")
	}
	if len(created.DaysOff) != 1 || created.DaysOff[0].String() != "2026-11-26" {
		t.Fatalf("POST create: unexpected days_off: %v", created.DaysOff)
	}

	terms := listTermsAuthed(t)
	if len(terms) < 2 {
		t.Fatalf("GET list after create: expected at least 2 terms, got %d", len(terms))
	}

	// Ordered by end_date ascending: Spring (2026-05-23) before Fall (2026-12-15)
	if terms[0].EndDate == nil || terms[1].EndDate == nil {
		t.Fatal("GET list: expected end_date on terms")
	}
	if terms[0].EndDate.Time.After(terms[1].EndDate.Time) {
		t.Fatalf("GET list: expected ascending end_date order, got %s then %s",
			terms[0].EndDate, terms[1].EndDate)
	}
}

func TestTermsHandlerPOSTCreateValidation(t *testing.T) {
	subTests := []struct {
		name         string
		body         interface{}
		expectedCode int
	}{
		{"missing end_date", map[string]interface{}{
			"name":     "Winter 2027",
			"days_off": []string{},
		}, http.StatusBadRequest},
		{"invalid end_date", map[string]interface{}{
			"name":     "Winter 2027",
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

			rr := sendTermRequestAuthed(http.MethodPost, bytes.NewReader(bodyBytes))
			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestTermsHandlerGETByID(t *testing.T) {
	active := getActiveTermAuthed(t)

	rr := sendTermRequestAuthedPath(http.MethodGet, fmt.Sprintf("/api/terms/%d", active.ID), nil)
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	var got termResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("GET by id: expected term JSON, got %q: %v", rr.Body.String(), err)
	}
	if got.ID != active.ID {
		t.Fatalf("GET by id: expected id %d, got %d", active.ID, got.ID)
	}
	if got.Name != active.Name {
		t.Fatalf("GET by id: expected name %q, got %q", active.Name, got.Name)
	}
}

func TestTermsHandlerGETByIDNotFound(t *testing.T) {
	rr := sendTermRequestAuthedPath(http.MethodGet, "/api/terms/999999999", nil)
	assertStatusAndBody(t, http.StatusNotFound, rr.Code, rr.Body)
}

func TestTermsHandlerPOSTActivate(t *testing.T) {
	// Ensure we have a second inactive term to activate.
	body := map[string]interface{}{
		"name":     "Winter 2027",
		"end_date": "2027-03-20",
		"days_off": []string{},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	rr := sendTermRequestAuthed(http.MethodPost, bytes.NewReader(bodyBytes))
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	var created termResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("POST create: expected term JSON, got %q: %v", rr.Body.String(), err)
	}
	if created.IsActive {
		t.Fatal("POST create: expected inactive term before activate")
	}

	previousActive := getActiveTermAuthed(t)

	rr = sendTermRequestAuthedPath(http.MethodPost, fmt.Sprintf("/api/terms/%d/activate", created.ID), nil)
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	terms := listTermsAuthed(t)
	var newActive, oldTerm *termResponse
	for i := range terms {
		term := &terms[i]
		if term.ID == created.ID {
			newActive = term
		}
		if term.ID == previousActive.ID {
			oldTerm = term
		}
	}
	if newActive == nil {
		t.Fatal("after activate: created term missing from list")
	}
	if !newActive.IsActive {
		t.Fatal("after activate: created term should be active")
	}
	if oldTerm == nil {
		t.Fatal("after activate: previous active term missing from list")
	}
	if oldTerm.IsActive {
		t.Fatal("after activate: previous active term should be inactive")
	}

	activeCount := 0
	for _, term := range terms {
		if term.IsActive {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Fatalf("after activate: expected exactly one active term, got %d", activeCount)
	}
}

func TestTermsHandlerPOSTActivateNotFound(t *testing.T) {
	rr := sendTermRequestAuthedPath(http.MethodPost, "/api/terms/999999999/activate", nil)
	assertStatusAndBody(t, http.StatusNotFound, rr.Code, rr.Body)
}

func TestParseTermsPath(t *testing.T) {
	subTests := []struct {
		path      string
		wantRoute termsRoute
		wantID    int64
		wantErr   bool
	}{
		{"/api/terms", termsRouteCollection, 0, false},
		{"/api/terms/", termsRouteCollection, 0, false},
		{"/api/terms/42", termsRouteByID, 42, false},
		{"/api/terms/42/activate", termsRouteActivate, 42, false},
		{"/api/terms/abc", 0, 0, true},
		{"/api/terms/0", 0, 0, true},
		{"/api/terms/42/entries", 0, 0, true},
	}

	for _, st := range subTests {
		t.Run(st.path, func(tt *testing.T) {
			route, id, err := parseTermsPath(st.path)
			if st.wantErr {
				if err == nil {
					tt.Fatal("expected error")
				}
				return
			}
			if err != nil {
				tt.Fatalf("unexpected error: %v", err)
			}
			if route != st.wantRoute || id != st.wantID {
				tt.Fatalf("got route=%v id=%d, want route=%v id=%d", route, id, st.wantRoute, st.wantID)
			}
		})
	}
}

func listTermsAuthed(t *testing.T) []termResponse {
	t.Helper()
	rr := sendTermRequestAuthed(http.MethodGet, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET terms: expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	var terms []termResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &terms); err != nil {
		t.Fatalf("GET terms: expected JSON array, got %q: %v", rr.Body.String(), err)
	}
	return terms
}

func getActiveTermAuthed(t *testing.T) termResponse {
	t.Helper()
	terms := listTermsAuthed(t)
	for _, term := range terms {
		if term.IsActive {
			return term
		}
	}
	t.Fatal("expected an active term")
	return termResponse{}
}

func sendTermRequestAuthed(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendTermRequest(method, body, testJWTTokenCookie)
}

func sendTermRequest(method string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendTermRequestPath(method, "/api/terms", body, auth)
}

func sendTermRequestAuthedPath(method, path string, body io.Reader) *httptest.ResponseRecorder {
	return sendTermRequestPath(method, path, body, testJWTTokenCookie)
}

func sendTermRequestPath(method, path string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendRequest(method, path, body, auth, TermsHandler)
}
