package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nishoof/flexi/backend/database"
	"github.com/nishoof/flexi/backend/util"
)

const tableBudgets = "flex_budgets"

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

	switch r.Method {
	case http.MethodGet:
		response, err = getBudget(userId)
	case http.MethodPut:
		err = updateBudget(r.Body, userId)
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

func getBudget(userId int64) ([]byte, error) {
	query := tableBudgets + "?user_id=eq." + fmt.Sprint(userId)
	responseBody, err := database.Request(http.MethodGet, query, nil)
	if err != nil {
		return nil, err
	}

	if len(responseBody) == 2 { // empty array: "[]"
		// No budget found for user, so create a default one
		err = createDefaultBudget(userId)
		if err != nil {
			return nil, err
		}
		return getBudget(userId)
	}

	return responseBody, nil
}

func updateBudget(body io.ReadCloser, userId int64) error {
	budget := Budget{}
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&budget)
	if err != nil {
		return errInvalidBudget
	}
	budget.UserId = userId

	valid := isValidBudget(budget)
	if !valid {
		return errInvalidBudget
	}

	budgetBytes, err := json.Marshal(budget)
	if err != nil {
		return fmt.Errorf("Failed to marshal budget data: %w", err)
	}
	budgetReader := bytes.NewReader(budgetBytes)

	_, err = database.Request(http.MethodPatch, tableBudgets+"?user_id=eq."+fmt.Sprint(userId), budgetReader)
	return err
}

func createDefaultBudget(userId int64) error {
	defaultBudget := Budget{
		UserId:   userId,
		Holidays: []*util.Date{},
	}

	budgetBytes, err := json.Marshal(defaultBudget)
	if err != nil {
		return fmt.Errorf("Failed to marshal default budget data: %w", err)
	}
	budgetReader := bytes.NewReader(budgetBytes)

	_, err = database.Request(http.MethodPost, tableBudgets, budgetReader)
	return err
}

func isValidBudget(budget Budget) bool {
	// UserId
	if budget.UserId <= 0 {
		return false
	}

	// Holidays
	for _, holiday := range budget.Holidays {
		if holiday == nil {
			return false
		}
	}

	return true
}
