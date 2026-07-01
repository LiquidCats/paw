-- name: CreateKey :execresult
INSERT INTO keys
(key_id, seed_fingerprint, alias, curve, algorithm, derivation_path, status, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING key_id, expires_at;

-- name: GetKey :one
SELECT *
FROM keys
WHERE key_id = $1
LIMIT 1;

-- name: GetAllKeys :many
SELECT *
FROM keys
ORDER BY created_at;

-- name: SetStatus :exec
UPDATE keys
SET status = $2
WHERE key_id = $1;

-- name: SetExpiration :exec
UPDATE keys
SET expires_at = $2
WHERE key_id = $1;
