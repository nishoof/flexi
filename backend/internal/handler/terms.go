package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/internal/database"
	"github.com/nishoof/flexi/backend/internal/repository"
	"github.com/nishoof/flexi/backend/internal/util"
)

var errInvalidTerm = errors.New("Invalid term data")
var errTermNotFound = errors.New("Term not found")

type termResponse struct {
	ID             int64        `json:"id,omitempty"`
	Name           string       `json:"name"`
	StartDate      *util.Date   `json:"start_date"`
	EndDate        *util.Date   `json:"end_date"`
	StartingAmount float64      `json:"starting_amount"`
	IsActive       bool         `json:"is_active"`
	DaysOff        []*util.Date `json:"days_off"`
}

type termUpdate struct {
	Name           string       `json:"name"`
	StartDate      *util.Date   `json:"start_date"`
	EndDate        *util.Date   `json:"end_date"`
	StartingAmount *float64     `json:"starting_amount"`
	DaysOff        []*util.Date `json:"days_off"`
}

var (
	defaultTermName       = "Spring 2026"
	defaultStartDate, _   = time.Parse("2006-01-02", "2026-01-26")
	defaultEndDate, _     = time.Parse("2006-01-02", "2026-05-23")
	defaultStartingAmount = 3010.0
)

func RegisterTerms(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/terms", ListTermsHandler)
	mux.HandleFunc("POST /api/terms", CreateTermHandler)
	mux.HandleFunc("PUT /api/terms", UpdateTermHandler)
	mux.HandleFunc("GET /api/terms/{id}", GetTermHandler)
	mux.HandleFunc("POST /api/terms/{id}/activate", ActivateTermHandler)
}

func ListTermsHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := authenticateTermRequest(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	queries, err := database.Queries(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	terms, err := queries.ListTerms(ctx, userId)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	// Create a default active term when the user has none
	if len(terms) == 0 {
		_, err := queries.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
			UserID: userId,
			Name:   defaultTermName,
			StartDate: pgtype.Date{
				Time:  defaultStartDate,
				Valid: true,
			},
			EndDate: pgtype.Date{
				Time:  defaultEndDate,
				Valid: true,
			},
			StartingAmount: defaultStartingAmount,
		})
		if err != nil {
			writeTermResult(w, nil, err)
			return
		}
		terms, err = queries.ListTerms(ctx, userId)
		if err != nil {
			writeTermResult(w, nil, err)
			return
		}
	}

	response, err := marshalTerms(ctx, queries, terms)
	writeTermResult(w, response, err)
}

func CreateTermHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := authenticateTermRequest(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	input := termUpdate{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeTermResult(w, nil, errInvalidTerm)
		return
	}
	if !isValidTermUpdate(input) {
		writeTermResult(w, nil, errInvalidTerm)
		return
	}

	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}
	defer tx.Rollback(ctx)

	term, err := qtx.CreateTerm(ctx, repository.CreateTermParams{
		UserID: userId,
		Name:   input.Name,
		StartDate: pgtype.Date{
			Time:  input.StartDate.Time,
			Valid: true,
		},
		EndDate: pgtype.Date{
			Time:  input.EndDate.Time,
			Valid: true,
		},
		StartingAmount: *input.StartingAmount,
	})
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	for _, dayOff := range input.DaysOff {
		if err := qtx.InsertDayOff(ctx, repository.InsertDayOffParams{
			TermID: term.ID,
			Date: pgtype.Date{
				Time:  dayOff.Time,
				Valid: true,
			},
		}); err != nil {
			writeTermResult(w, nil, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeTermResult(w, nil, err)
		return
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}
	response, err := marshalTerm(ctx, queries, term)
	writeTermResult(w, response, err)
}

func UpdateTermHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := authenticateTermRequest(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	update := termUpdate{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&update); err != nil {
		writeTermResult(w, nil, errInvalidTerm)
		return
	}
	if !isValidTermUpdate(update) {
		writeTermResult(w, nil, errInvalidTerm)
		return
	}

	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}
	defer tx.Rollback(ctx)

	term, err := qtx.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
		UserID: userId,
		Name:   update.Name,
		StartDate: pgtype.Date{
			Time:  update.StartDate.Time,
			Valid: true,
		},
		EndDate: pgtype.Date{
			Time:  update.EndDate.Time,
			Valid: true,
		},
		StartingAmount: *update.StartingAmount,
	})
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	if err := qtx.UpdateActiveTerm(ctx, repository.UpdateActiveTermParams{
		ID:   term.ID,
		Name: update.Name,
		StartDate: pgtype.Date{
			Time:  update.StartDate.Time,
			Valid: true,
		},
		EndDate: pgtype.Date{
			Time:  update.EndDate.Time,
			Valid: true,
		},
		StartingAmount: *update.StartingAmount,
	}); err != nil {
		writeTermResult(w, nil, err)
		return
	}

	if err := qtx.DeleteDaysOffByTerm(ctx, term.ID); err != nil {
		writeTermResult(w, nil, err)
		return
	}

	for _, dayOff := range update.DaysOff {
		if err := qtx.InsertDayOff(ctx, repository.InsertDayOffParams{
			TermID: term.ID,
			Date: pgtype.Date{
				Time:  dayOff.Time,
				Valid: true,
			},
		}); err != nil {
			writeTermResult(w, nil, err)
			return
		}
	}

	writeTermResult(w, nil, tx.Commit(ctx))
}

func GetTermHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := authenticateTermRequest(w, r)
	if !ok {
		return
	}
	termID, err := parseTermID(r)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	ctx := r.Context()

	queries, err := database.Queries(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	term, err := queries.GetTermByID(ctx, repository.GetTermByIDParams{
		ID:     termID,
		UserID: userId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeTermResult(w, nil, errTermNotFound)
		return
	}
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	response, err := marshalTerm(ctx, queries, term)
	writeTermResult(w, response, err)
}

func ActivateTermHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := authenticateTermRequest(w, r)
	if !ok {
		return
	}
	termID, err := parseTermID(r)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	ctx := r.Context()

	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}
	defer tx.Rollback(ctx)

	if err := qtx.DeactivateTermsByUser(ctx, userId); err != nil {
		writeTermResult(w, nil, err)
		return
	}

	_, err = qtx.ActivateTerm(ctx, repository.ActivateTermParams{
		ID:     termID,
		UserID: userId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeTermResult(w, nil, errTermNotFound)
		return
	}
	if err != nil {
		writeTermResult(w, nil, err)
		return
	}

	writeTermResult(w, nil, tx.Commit(ctx))
}

func authenticateTermRequest(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userId, err := util.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return 0, false
	}
	return userId, true
}

func parseTermID(r *http.Request) (int64, error) {
	termID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || termID <= 0 {
		return 0, errors.New("Not Found")
	}
	return termID, nil
}

func writeTermResult(w http.ResponseWriter, response []byte, err error) {
	if errors.Is(err, errInvalidTerm) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, errTermNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

// getOrCreateTerm ensures an active term exists (used by test setup).
func getOrCreateTerm(ctx context.Context, userId int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	term, err := queries.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
		UserID: userId,
		Name:   defaultTermName,
		StartDate: pgtype.Date{
			Time:  defaultStartDate,
			Valid: true,
		},
		EndDate: pgtype.Date{
			Time:  defaultEndDate,
			Valid: true,
		},
		StartingAmount: defaultStartingAmount,
	})
	if err != nil {
		return nil, err
	}

	return marshalTerm(ctx, queries, term)
}

func marshalTerms(ctx context.Context, queries *repository.Queries, terms []repository.Term) ([]byte, error) {
	responses := make([]termResponse, 0, len(terms))
	for _, term := range terms {
		resp, err := termToResponse(ctx, queries, term)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return json.Marshal(responses)
}

func marshalTerm(ctx context.Context, queries *repository.Queries, term repository.Term) ([]byte, error) {
	resp, err := termToResponse(ctx, queries, term)
	if err != nil {
		return nil, err
	}
	return json.Marshal(resp)
}

func termToResponse(ctx context.Context, queries *repository.Queries, term repository.Term) (termResponse, error) {
	daysOff, err := queries.ListDaysOff(ctx, term.ID)
	if err != nil {
		return termResponse{}, err
	}

	startDate, err := util.NewDate(term.StartDate.Time.Format("2006-01-02"))
	if err != nil {
		return termResponse{}, err
	}

	endDate, err := util.NewDate(term.EndDate.Time.Format("2006-01-02"))
	if err != nil {
		return termResponse{}, err
	}

	daysOffDates := make([]*util.Date, 0, len(daysOff))
	for _, date := range daysOff {
		d, err := util.NewDate(date.Time.Format("2006-01-02"))
		if err != nil {
			return termResponse{}, err
		}
		daysOffDates = append(daysOffDates, d)
	}

	return termResponse{
		ID:             term.ID,
		Name:           term.Name,
		StartDate:      startDate,
		EndDate:        endDate,
		StartingAmount: term.StartingAmount,
		IsActive:       term.IsActive,
		DaysOff:        daysOffDates,
	}, nil
}

func isValidTermUpdate(update termUpdate) bool {
	if update.StartDate == nil || update.EndDate == nil {
		return false
	}
	if update.StartingAmount == nil || *update.StartingAmount < 0 {
		return false
	}
	if update.StartDate.Time.After(update.EndDate.Time) {
		return false
	}
	for _, dayOff := range update.DaysOff {
		if dayOff == nil {
			return false
		}
		if dayOff.Time.Before(update.StartDate.Time) || dayOff.Time.After(update.EndDate.Time) {
			return false
		}
	}
	return true
}
