-- name: GetStateByKey :one
select key, value, created_at, updated_at from state where key = $1;

-- name: SetState :exec
INSERT INTO state (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = current_timestamp;