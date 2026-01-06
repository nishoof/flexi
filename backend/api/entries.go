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

const table = "flex_entries"

var errInvalidEntry = errors.New("Invalid entry data")

type entry struct {
	AmountRemaining float64 `json:"amount_remaining"`
	Date            string  `json:"date"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	isOptionsRequest := util.HandleCORS(w, r)
	if isOptionsRequest {
		return
	}

	var body []byte
	var err error

	switch r.Method {
	case http.MethodGet:
		body, err = database.Request(http.MethodGet, table, nil)
	case http.MethodPost:
		body, err = addEntry(r.Body)
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

	w.Write(body)
}

func addEntry(body io.ReadCloser) ([]byte, error) {
	e := entry{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&e)
	if err != nil {
		return nil, errInvalidEntry // invalid JSON
	}
	valid := validateEntry(e)
	if !valid {
		return nil, errInvalidEntry
	}

	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal entry data: %w", err)
	}

	return database.Request(http.MethodPost, table, bytes.NewReader(data))
}

func validateEntry(e entry) bool {
	if e.AmountRemaining < 0 {
		return false
	}
	if e.Date == "" {
		return false
	}
	const layout = "2006-01-02" // see https://golang.org/pkg/time/#Time.Format
	_, err := time.Parse(layout, e.Date)
	if err != nil {
		return false
	}
	return true
}
