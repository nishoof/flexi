package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/util"
)

const tableEntries = "flex_entries"

var errInvalidEntry = errors.New("Invalid entry data")
var errUnexpectedDbResponse = errors.New("Unexpected response from database")

// Entry represents a flexi entry (how much flexi a user has remaining at a given date).
// Pointers are used to distinguish between missing and zero values
type entry struct {
	UserId          int64    `json:"user_id"`
	AmountRemaining *float64 `json:"amount_remaining"`
	Date            *string  `json:"date"`
}

func EntriesHandler(w http.ResponseWriter, r *http.Request) {
	isOptionsRequest := util.HandleCORS(w, r)
	if isOptionsRequest {
		return
	}

	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jwt := cookie.Value
	valid := util.VerifyJWT(jwt)
	if !valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userId, err := util.GetUserIdFromJWT(jwt)
	if err != nil {
		fmt.Println("Error extracting user ID from JWT:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var response []byte

	switch r.Method {
	case http.MethodGet:
		response, err = getEntries(userId)
	case http.MethodPost:
		err = createEntry(r.Body, userId)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if errors.Is(err, errInvalidEntry) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)

}

func getEntries(userId int64) ([]byte, error) {
	query := tableEntries + "?user_id=eq." + fmt.Sprint(userId) + "&order=date.desc"
	responseBody, err := database.Request(http.MethodGet, query, nil)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
}

func createEntry(body io.ReadCloser, userId int64) error {
	entry := entry{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&entry)
	if err != nil {
		return errInvalidEntry // invalid JSON
	}
	entry.UserId = userId

	valid := validateEntry(entry)
	if !valid {
		return errInvalidEntry
	}

	entryBytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("Failed to marshal entry data: %w", err)
	}
	entryReader := bytes.NewReader(entryBytes)

	dbResponse, err := database.Request(http.MethodPost, tableEntries, entryReader)
	if err != nil {
		return err
	}
	if len(dbResponse) != 0 {
		return fmt.Errorf("%w: %s", errUnexpectedDbResponse, string(dbResponse))
	}

	return nil
}

func validateEntry(entry entry) bool {
	// UserId
	if entry.UserId <= 0 {
		return false
	}

	// AmountRemaining
	if entry.AmountRemaining == nil {
		return false
	}
	if *entry.AmountRemaining < 0 {
		return false
	}

	// Date
	if entry.Date == nil {
		return false
	}
	if *entry.Date == "" {
		return false
	}
	const layout = "2006-01-02" // see https://golang.org/pkg/time/#Time.Format
	_, err := time.Parse(layout, *entry.Date)
	if err != nil {
		return false
	}
	return true
}
