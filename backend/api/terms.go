package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/repository"
	"github.com/nishoof/flexi/backend/util"
)

var errInvalidTerm = errors.New("Invalid term data")

type termResponse struct {
	ID       int64        `json:"id,omitempty"`
	Name     string       `json:"name"`
	EndDate  *util.Date   `json:"end_date"`
	IsActive bool         `json:"is_active"`
	DaysOff  []*util.Date `json:"days_off"`
}

type termUpdate struct {
	Name    string       `json:"name"`
	EndDate *util.Date   `json:"end_date"`
	DaysOff []*util.Date `json:"days_off"`
}

func TermsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := util.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var response []byte
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		response, err = getOrCreateTerm(ctx, userId)
	case http.MethodPut:
		err = updateTerm(ctx, r.Body, userId)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if errors.Is(err, errInvalidTerm) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func getOrCreateTerm(ctx context.Context, userId int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	defaultTermName := "Spring 2026"
	defaultEndDate, _ := time.Parse("2006-01-02", "2026-05-23")

	term, err := queries.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
		UserID: userId,
		Name:   defaultTermName,
		EndDate: pgtype.Date{
			Time:  defaultEndDate,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	daysOff, err := queries.ListDaysOff(ctx, term.ID)
	if err != nil {
		return nil, err
	}

	endDate, err := util.NewDate(term.EndDate.Time.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	daysOffDates := make([]*util.Date, 0, len(daysOff))
	for _, date := range daysOff {
		d, err := util.NewDate(date.Time.Format("2006-01-02"))
		if err != nil {
			return nil, err
		}
		daysOffDates = append(daysOffDates, d)
	}

	response := termResponse{
		ID:       term.ID,
		Name:     term.Name,
		EndDate:  endDate,
		IsActive: term.IsActive,
		DaysOff:  daysOffDates,
	}
	return json.Marshal(response)
}

func updateTerm(ctx context.Context, body io.ReadCloser, userId int64) error {
	update := termUpdate{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&update); err != nil {
		return errInvalidTerm
	}

	if !isValidTermUpdate(update) {
		return errInvalidTerm
	}

	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	term, err := qtx.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
		UserID: userId,
		Name:   update.Name,
		EndDate: pgtype.Date{
			Time:  update.EndDate.Time,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	if err := qtx.UpdateActiveTerm(ctx, repository.UpdateActiveTermParams{
		ID:   term.ID,
		Name: update.Name,
		EndDate: pgtype.Date{
			Time:  update.EndDate.Time,
			Valid: true,
		},
	}); err != nil {
		return err
	}

	if err := qtx.DeleteDaysOffByTerm(ctx, term.ID); err != nil {
		return err
	}

	for _, dayOff := range update.DaysOff {
		if err := qtx.InsertDayOff(ctx, repository.InsertDayOffParams{
			TermID: term.ID,
			Date: pgtype.Date{
				Time:  dayOff.Time,
				Valid: true,
			},
		}); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func isValidTermUpdate(update termUpdate) bool {
	if update.EndDate == nil {
		return false
	}
	for _, dayOff := range update.DaysOff {
		if dayOff == nil {
			return false
		}
	}
	return true
}
