-- name: ListEntries :many
SELECT e.amount_remaining, e.date
FROM app.entries e
JOIN app.terms t ON t.id = e.term_id
WHERE t.user_id = $1 AND t.is_active = true
ORDER BY e.date DESC;

-- name: CreateEntry :execrows
INSERT INTO app.entries (term_id, amount_remaining, date)
SELECT t.id, $2, $3
FROM app.terms t
WHERE t.user_id = $1 AND t.is_active = true
ON CONFLICT (term_id, date) DO NOTHING;

-- name: DeleteEntriesByUser :exec
DELETE FROM app.entries e
USING app.terms t
WHERE e.term_id = t.id
  AND t.user_id = $1;
