package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/nishoof/flexi/backend/database"
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
	pool, err := database.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var holidays []byte
	err = pool.QueryRow(ctx,
		`SELECT holidays FROM app.budgets WHERE user_id = $1`, userId,
	).Scan(&holidays)

	if errors.Is(err, pgx.ErrNoRows) {
		if err := createDefaultBudget(ctx, userId); err != nil {
			return nil, err
		}
		return getBudget(ctx, userId)
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal([]map[string]json.RawMessage{
		{"holidays": holidays},
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

	pool, err := database.Pool(ctx)
	if err != nil {
		return err
	}

	holidays, err := json.Marshal(budget.Holidays)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO app.budgets (user_id, holidays)
		 VALUES ($1, $2::jsonb)
		 ON CONFLICT (user_id) DO UPDATE SET holidays = EXCLUDED.holidays`,
		userId, string(holidays),
	)
	if err != nil {
		fmt.Println("Error updating budget in database:", err)
	}
	return err
}

func createDefaultBudget(ctx context.Context, userId int64) error {
	pool, err := database.Pool(ctx)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO app.budgets (user_id, holidays)
		 VALUES ($1, '[]'::jsonb)
		 ON CONFLICT (user_id) DO NOTHING`,
		userId,
	)
	return err
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
