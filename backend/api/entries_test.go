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
	"github.com/nishoof/flexi/backend/util"
)

func TestEntriesHandler(t *testing.T) {
	rr := sendEntriesPostRequest(t, 1000.50, testEntryDate)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST entry: expected status %d, got %d, body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	rr = sendEntriesRequestAuthed(http.MethodGet, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET entries: expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var entries []entry
	if err := json.Unmarshal(rr.Body.Bytes(), &entries); err != nil {
		t.Fatalf("Failed to unmarshal entries: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.Date != nil && e.Date.String() == testEntryDate {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Expected to find entry with date " + testEntryDate)
	}
}

func TestEntriesHandlerAuth(t *testing.T) {
	subTests := []struct {
		name         string
		method       string
		authToken    string
		expectedCode int
	}{
		{"GET without auth token", http.MethodGet, "", http.StatusUnauthorized},
		{"POST without auth token", http.MethodPost, "", http.StatusUnauthorized},
		{"GET with invalid auth token", http.MethodGet, "invalid-token", http.StatusUnauthorized},
		{"POST with invalid auth token", http.MethodPost, "invalid-token", http.StatusUnauthorized},
	}

	for _, st := range subTests {
		t.Run(st.name, func(tt *testing.T) {
			var cookie *http.Cookie
			if st.authToken != "" {
				cookie = &http.Cookie{Name: "auth_token", Value: st.authToken}
			}
			rr := sendEntriesRequest(st.method, nil, cookie)

			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestEntriesHandlerMethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodPatch, http.MethodDelete, http.MethodPut}
	for _, method := range methods {
		t.Run(method, func(tt *testing.T) {
			rr := sendEntriesRequestAuthed(method, nil)
			assertStatusAndBody(tt, http.StatusMethodNotAllowed, rr.Code, rr.Body)
		})
	}
}

func TestEntriesHandlerValidation(t *testing.T) {
	subTests := []struct {
		name         string
		body         interface{}
		expectedCode int
	}{
		{"missing amount remaining", map[string]interface{}{"date": testEntryDate}, http.StatusBadRequest},
		{"negative amount remaining", map[string]interface{}{"amount_remaining": -1.0, "date": testEntryDate}, http.StatusBadRequest},
		{"missing date", map[string]interface{}{"amount_remaining": 100.0}, http.StatusBadRequest},
		{"invalid JSON", "not-json", http.StatusBadRequest},
		// {"valid entry", map[string]interface{}{"amount_remaining": 67.0, "date": testEntryDate}, http.StatusOK},
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

			rr := sendEntriesRequestAuthed(http.MethodPost, bytes.NewReader(bodyBytes))
			if st.expectedCode == http.StatusOK {
				registerEntriesCleanup(tt, testUserId)
			}
			assertStatusAndBody(tt, st.expectedCode, rr.Code, rr.Body)
		})
	}
}

func TestEntriesHandlerDuplicateEntry(t *testing.T) {
	rr := sendEntriesPostRequest(t, 1000.50, testEntryDate)
	assertStatusAndBody(t, http.StatusOK, rr.Code, rr.Body)

	rr2 := sendEntriesPostRequest(t, 999.99, testEntryDate)
	assertStatusAndBody(t, http.StatusBadRequest, rr2.Code, rr2.Body)
}

func sendEntriesPostRequest(t *testing.T, amount float64, dateStr string) *httptest.ResponseRecorder {
	entryDate, err := util.NewDate(dateStr)
	if err != nil {
		t.Fatalf("Failed to create date: %v", err)
	}
	body := entry{AmountRemaining: &amount, Date: entryDate}
	bodyBytes, _ := json.Marshal(body)

	rr := sendEntriesRequestAuthed(http.MethodPost, bytes.NewReader(bodyBytes))

	registerEntriesCleanup(t, testUserId)

	return rr
}

func sendEntriesRequestAuthed(method string, body io.Reader) *httptest.ResponseRecorder {
	return sendEntriesRequest(method, body, testJWTTokenCookie)
}

func sendEntriesRequest(method string, body io.Reader, auth *http.Cookie) *httptest.ResponseRecorder {
	return sendRequest(method, "/api/entries", body, auth, EntriesHandler)
}

func registerEntriesCleanup(t *testing.T, userId int64) {
	pool, err := database.Pool(context.Background())
	if err != nil {
		t.Fatalf("Failed to get database pool: %v", err)
	}

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(),
			`DELETE FROM flex_entries
			 WHERE user_id=$1`,
			userId)
		if err != nil {
			t.Fatalf("Failed to cleanup entries: %v", err)
		}
	})
}
