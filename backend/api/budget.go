package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/repository"
	"github.com/nishoof/flexi/backend/util"
)

var errInvalidBudget = errors.New("Invalid budget data")

// Budget represents a user's Budget settings, including holidays (dates where the user does not plan to spend).
type Budget struct {
	UserId   int64        `json:"user_id"`
	Holidays []*util.Date `json:"holidays"`
}

func BudgetHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := util.AuthenticateUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var response []byte
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		response, err = getBudget(ctx, userId)
	case http.MethodPut:
		err = updateBudget(ctx, r.Body, userId)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if errors.Is(err, errInvalidBudget) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func getBudget(ctx context.Context, userId int64) ([]byte, error) {
	queries, err := database.Queries(ctx)
	if err != nil {
		return nil, err
	}

	holidays, err := queries.GetHolidays(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return json.Marshal([]map[string]any{
			{"holidays": nil},
		})
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal([]map[string]json.RawMessage{
		{"holidays": json.RawMessage(holidays)},
	})
}

func updateBudget(ctx context.Context, body io.ReadCloser, userId int64) error {
	budget := Budget{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&budget); err != nil {
		return errInvalidBudget
	}
	budget.UserId = userId

	if !isValidBudget(budget) {
		return errInvalidBudget
	}

	holidays, err := json.Marshal(budget.Holidays)
	if err != nil {
		return err
	}

	queries, err := database.Queries(ctx)
	if err != nil {
		return err
	}

	return queries.UpsertBudget(ctx, repository.UpsertBudgetParams{
		UserID:   userId,
		Holidays: string(holidays),
	})
}

func isValidBudget(budget Budget) bool {
	if budget.UserId <= 0 {
		return false
	}
	for _, holiday := range budget.Holidays {
		if holiday == nil {
			return false
		}
	}
	return true
}
