package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/internal/apierr"
	"github.com/nishoof/flexi/backend/internal/database"
	"github.com/nishoof/flexi/backend/internal/repository"
	"github.com/nishoof/flexi/backend/internal/util"
)

// Entry represents a flexi entry (how much flexi a user has remaining at a given date).
// Pointers are used to distinguish between missing and zero values
type entry struct {
	UserId          int64      `json:"user_id"`
	AmountRemaining *float64   `json:"amount_remaining"`
	Date            *util.Date `json:"date"`
}

func RegisterEntries(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/entries", Handle(withAuth(ListEntriesHandler)))
	mux.HandleFunc("POST /api/entries", Handle(withAuth(CreateEntryHandler)))
}

func ListEntriesHandler(w http.ResponseWriter, r *http.Request) *apierr.Error {
	userId := userID(r)
	ctx := r.Context()

	queries, err := database.Queries(ctx)
	if err != nil {
		return apierr.Internal(err)
	}

	rows, err := queries.ListEntries(ctx, userId)
	if err != nil {
		return apierr.Internal(err)
	}

	response := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		date, err := util.NewDate(row.Date.Time.Format("2006-01-02"))
		if err != nil {
			return apierr.Internal(err)
		}
		amount := row.AmountRemaining
		response = append(response, map[string]any{
			"amount_remaining": &amount,
			"date":             date,
		})
	}
	body, err := json.Marshal(response)
	if err != nil {
		return apierr.Internal(err)
	}
	w.Write(body)
	return nil
}

func CreateEntryHandler(w http.ResponseWriter, r *http.Request) *apierr.Error {
	userId := userID(r)
	ctx := r.Context()

	e := entry{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&e); err != nil {
		return apierr.BadRequest("Invalid entry data")
	}
	e.UserId = userId

	if !isValidEntry(e) {
		return apierr.BadRequest("Invalid entry data")
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		return apierr.Internal(err)
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
		return apierr.Internal(err)
	}
	if rowsAffected == 0 {
		return apierr.BadRequest(fmt.Sprintf("An entry for the date %s already exists", e.Date.String()))
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
