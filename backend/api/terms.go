package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/repository"
	"github.com/nishoof/flexi/backend/util"
)

var errInvalidTerm = errors.New("Invalid term data")
var errTermNotFound = errors.New("Term not found")

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

type termsRoute int

const (
	termsRouteCollection termsRoute = iota // /api/terms
	termsRouteByID                         // /api/terms/:id
	termsRouteActivate                     // /api/terms/:id/activate
)

var (
	defaultTermName   = "Spring 2026"
	defaultEndDate, _ = time.Parse("2006-01-02", "2026-05-23")
)

func TermsHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := util.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	route, termID, err := parseTermsPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var response []byte

	switch route {
	case termsRouteCollection:
		response, err = termsRouteCollectionHandler(w, r, userId)
	case termsRouteByID:
		response, err = termsRouteByIdHandler(w, r, userId, termID)
	case termsRouteActivate:
		err = termsRouteActivateHandler(w, r, userId, termID)
	}

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

func termsRouteCollectionHandler(w http.ResponseWriter, r *http.Request, userId int64) ([]byte, error) {
	var response []byte
	var err error
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		response, err = listTerms(ctx, userId)
	case http.MethodPut:
		err = updateTerm(ctx, r.Body, userId)
	case http.MethodPost:
		response, err = createTerm(ctx, r.Body, userId)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return nil, nil
	}

	return response, err
}

func termsRouteByIdHandler(w http.ResponseWriter, r *http.Request, userId, termID int64) ([]byte, error) {
	var response []byte
	var err error
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		response, err = getTermByID(ctx, userId, termID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return nil, nil
	}

	return response, err
}

func termsRouteActivateHandler(w http.ResponseWriter, r *http.Request, userId, termID int64) error {
	var err error
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		err = activateTerm(ctx, userId, termID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return nil
	}

	return err
}

// parseTermsPath maps /api/terms, /api/terms/:id, and /api/terms/:id/activate.
func parseTermsPath(path string) (termsRoute, int64, error) {
	path = strings.TrimPrefix(path, "/api/terms")
	path = strings.Trim(path, "/")
	if path == "" {
		return termsRouteCollection, 0, nil
	}

	parts := strings.Split(path, "/")
	termID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || termID <= 0 {
		return 0, 0, errors.New("Not Found")
	}

	switch len(parts) {
	case 1:
		return termsRouteByID, termID, nil
	case 2:
		if parts[1] == "activate" {
			return termsRouteActivate, termID, nil
		}
	}
	return 0, 0, errors.New("Not Found")
}

func listTerms(ctx context.Context, userId int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	terms, err := queries.ListTerms(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Create a default active term when the user has none
	if len(terms) == 0 {
		_, err := queries.GetOrCreateActiveTerm(ctx, repository.GetOrCreateActiveTermParams{
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
		terms, err = queries.ListTerms(ctx, userId)
		if err != nil {
			return nil, err
		}
	}

	return marshalTerms(ctx, queries, terms)
}

func createTerm(ctx context.Context, body io.ReadCloser, userId int64) ([]byte, error) {
	input := termUpdate{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return nil, errInvalidTerm
	}
	if !isValidTermUpdate(input) {
		return nil, errInvalidTerm
	}

	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	term, err := qtx.CreateTerm(ctx, repository.CreateTermParams{
		UserID: userId,
		Name:   input.Name,
		EndDate: pgtype.Date{
			Time:  input.EndDate.Time,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, dayOff := range input.DaysOff {
		if err := qtx.InsertDayOff(ctx, repository.InsertDayOffParams{
			TermID: term.ID,
			Date: pgtype.Date{
				Time:  dayOff.Time,
				Valid: true,
			},
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}
	return marshalTerm(ctx, queries, term)
}

func getTermByID(ctx context.Context, userId, termID int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	term, err := queries.GetTermByID(ctx, repository.GetTermByIDParams{
		ID:     termID,
		UserID: userId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errTermNotFound
	}
	if err != nil {
		return nil, err
	}

	return marshalTerm(ctx, queries, term)
}

func activateTerm(ctx context.Context, userId, termID int64) error {
	qtx, tx, err := database.QueriesWithTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := qtx.DeactivateTermsByUser(ctx, userId); err != nil {
		return err
	}

	_, err = qtx.ActivateTerm(ctx, repository.ActivateTermParams{
		ID:     termID,
		UserID: userId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return errTermNotFound
	}
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
		EndDate: pgtype.Date{
			Time:  defaultEndDate,
			Valid: true,
		},
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
		ID:       term.ID,
		Name:     term.Name,
		EndDate:  endDate,
		IsActive: term.IsActive,
		DaysOff:  daysOffDates,
	}, nil
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
