-- name: ListEntries :many
SELECT amount_remaining, date
FROM app.entries
WHERE user_id = $1
ORDER BY date DESC;

-- name: CreateEntry :execrows
INSERT INTO app.entries (user_id, amount_remaining, date)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, date) DO NOTHING;

-- name: DeleteEntriesByUser :exec
DELETE FROM app.entries
WHERE user_id = $1;
