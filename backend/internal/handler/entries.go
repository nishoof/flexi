package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/internal/database"
	"github.com/nishoof/flexi/backend/internal/repository"
	"github.com/nishoof/flexi/backend/internal/util"
)

var errInvalidEntry = errors.New("Invalid entry data")

// Entry represents a flexi entry (how much flexi a user has remaining at a given date).
// Pointers are used to distinguish between missing and zero values
type entry struct {
	UserId          int64      `json:"user_id"`
	AmountRemaining *float64   `json:"amount_remaining"`
	Date            *util.Date `json:"date"`
}

func RegisterEntries(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/entries", withAuth(ListEntriesHandler))
	mux.HandleFunc("POST /api/entries", withAuth(CreateEntryHandler))
}

func ListEntriesHandler(w http.ResponseWriter, r *http.Request) {
	userId := userID(r)
	ctx := r.Context()

	queries, err := database.Queries(ctx)
	if err != nil {
		writeEntryResult(w, nil, err)
		return
	}

	rows, err := queries.ListEntries(ctx, userId)
	if err != nil {
		writeEntryResult(w, nil, err)
		return
	}

	response := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		date, err := util.NewDate(row.Date.Time.Format("2006-01-02"))
		if err != nil {
			writeEntryResult(w, nil, err)
			return
		}
		amount := row.AmountRemaining
		response = append(response, map[string]any{
			"amount_remaining": &amount,
			"date":             date,
		})
	}
	body, err := json.Marshal(response)
	writeEntryResult(w, body, err)
}

func CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	userId := userID(r)
	ctx := r.Context()

	e := entry{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&e); err != nil {
		writeEntryResult(w, nil, errInvalidEntry)
		return
	}
	e.UserId = userId

	if !isValidEntry(e) {
		writeEntryResult(w, nil, errInvalidEntry)
		return
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		writeEntryResult(w, nil, err)
		return
	}

	rowsAffected, err := queries.CreateEntry(ctx, repository.CreateEntryParams{
		UserID:          userId,
		AmountRemaining: *e.AmountRemaining,
		Date: pgtype.Date{
			Time:  e.Date.Time,
			Valid: true,
		},
	})
	if err != nil {
		writeEntryResult(w, nil, err)
		return
	}
	if rowsAffected == 0 {
		writeEntryResult(w, nil, fmt.Errorf("%w: An entry for the date %s already exists", errInvalidEntry, e.Date.String()))
		return
	}
	writeEntryResult(w, nil, nil)
}

func writeEntryResult(w http.ResponseWriter, response []byte, err error) {
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

func isValidEntry(entry entry) bool {
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
	return entry.Date != nil
}
