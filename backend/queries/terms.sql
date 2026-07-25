-- name: GetOrCreateActiveTerm :one
INSERT INTO app.terms (user_id, name, start_date, end_date, starting_amount, is_active)
VALUES ($1, $2, $3, $4, $5, true)
ON CONFLICT (user_id) WHERE is_active = true DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING id, user_id, name, end_date, is_active, created_at, start_date, starting_amount;

-- name: ListTerms :many
SELECT id, user_id, name, end_date, is_active, created_at, start_date, starting_amount
FROM app.terms
WHERE user_id = $1
ORDER BY end_date;

-- name: GetTermByID :one
SELECT id, user_id, name, end_date, is_active, created_at, start_date, starting_amount
FROM app.terms
WHERE id = $1 AND user_id = $2;

-- name: CreateTerm :one
INSERT INTO app.terms (user_id, name, start_date, end_date, starting_amount, is_active)
VALUES ($1, $2, $3, $4, $5, false)
RETURNING id, user_id, name, end_date, is_active, created_at, start_date, starting_amount;

-- name: DeactivateTermsByUser :exec
UPDATE app.terms
SET is_active = false
WHERE user_id = $1 AND is_active = true;

-- name: ActivateTerm :one
UPDATE app.terms
SET is_active = true
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, end_date, is_active, created_at, start_date, starting_amount;

-- name: UpdateActiveTerm :exec
UPDATE app.terms
SET name = $2, start_date = $3, end_date = $4, starting_amount = $5
WHERE id = $1;

-- name: ListDaysOff :many
SELECT date
FROM app.term_days_off
WHERE term_id = $1
ORDER BY date;

-- name: DeleteDaysOffByTerm :exec
DELETE FROM app.term_days_off
WHERE term_id = $1;

-- name: InsertDayOff :exec
INSERT INTO app.term_days_off (term_id, date)
VALUES ($1, $2);

-- name: DeleteTermsByUser :exec
DELETE FROM app.terms
WHERE user_id = $1;
