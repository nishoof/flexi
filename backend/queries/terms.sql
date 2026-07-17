-- name: GetOrCreateActiveTerm :one
INSERT INTO app.terms (user_id, name, end_date, is_active)
VALUES ($1, $2, $3, true)
ON CONFLICT (user_id) WHERE is_active = true DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING id, user_id, name, end_date, is_active, created_at;

-- name: UpdateActiveTerm :exec
UPDATE app.terms
SET name = $2, end_date = $3
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

-- name: DeleteActiveTermByUser :exec
DELETE FROM app.terms
WHERE user_id = $1 AND is_active = true;
