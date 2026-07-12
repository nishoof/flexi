-- name: GetHolidays :one
SELECT holidays
FROM app.budgets
WHERE user_id = $1;

-- name: UpsertBudget :exec
INSERT INTO app.budgets (user_id, holidays)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET holidays = EXCLUDED.holidays;

-- name: DeleteBudgetByUser :exec
DELETE FROM app.budgets
WHERE user_id = $1;