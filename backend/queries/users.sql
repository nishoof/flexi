-- name: GetOrCreateUser :one
INSERT INTO app.users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING id;

-- name: DeleteUser :exec
DELETE FROM app.users
WHERE id = $1;
