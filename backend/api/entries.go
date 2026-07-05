package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nishoof/flexi/backend/database"
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
	pool, err := database.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx,
		`SELECT amount_remaining, date
		 FROM app.entries
		 WHERE user_id = $1
		 ORDER BY date DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	response := make([]map[string]any, 0, 100)
	for rows.Next() {
		var scannedAmount float64
		var scannedDate time.Time
		if err := rows.Scan(&scannedAmount, &scannedDate); err != nil {
			return nil, err
		}
		date, err := util.NewDate(scannedDate.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		response = append(response, map[string]any{
			"amount_remaining": &scannedAmount,
			"date":             date,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
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

	pool, err := database.Pool(ctx)
	if err != nil {
		return err
	}

	tag, err := pool.Exec(ctx,
		`INSERT INTO app.entries (user_id, amount_remaining, date)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, date) DO NOTHING`,
		userId, *e.AmountRemaining, e.Date.String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
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
