package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/repository"
	"github.com/nishoof/flexi/backend/util"
)

var errInvalidEntry = errors.New("Invalid entry data")

// Entry represents a flexi entry (how much flexi a user has remaining at a given date).
// Pointers are used to distinguish between missing and zero values
type entry struct {
	UserId          int64      `json:"user_id"`
	AmountRemaining *float64   `json:"amount_remaining"`
	Date            *util.Date `json:"date"`
}

func EntriesHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := util.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var response []byte
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		response, err = getEntries(ctx, userId)
	case http.MethodPost:
		err = createEntry(ctx, r.Body, userId)
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

func getEntries(ctx context.Context, userId int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := queries.ListEntries(ctx, userId)
	if err != nil {
		return nil, err
	}

	response := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		date, err := util.NewDate(row.Date.Time.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		amount := row.AmountRemaining
		response = append(response, map[string]any{
			"amount_remaining": &amount,
			"date":             date,
		})
	}
	return json.Marshal(response)
}

func createEntry(ctx context.Context, body io.ReadCloser, userId int64) error {
	e := entry{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&e); err != nil {
		return errInvalidEntry
	}
	e.UserId = userId

	if !isValidEntry(e) {
		return errInvalidEntry
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		return err
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
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: An entry for the date %s already exists", errInvalidEntry, e.Date.String())
	}
	return nil
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
