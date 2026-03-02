-- name: GetUserByID :one
SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
FROM users
WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, role, building_id, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET username = $2, email = $3, password_hash = $4, role = $5, building_id = $6, status = $7
WHERE id = $1
RETURNING updated_at;

-- name: ListUsers :many
SELECT id, username, email, password_hash, role, building_id, status, created_at, updated_at
FROM users
WHERE (sqlc.narg('filter_role')::varchar IS NULL OR role = sqlc.narg('filter_role'))
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR building_id = sqlc.narg('filter_building_id'))
  AND (sqlc.narg('filter_status')::varchar IS NULL OR status = sqlc.narg('filter_status'))
ORDER BY created_at DESC
LIMIT sqlc.narg('max_results')
OFFSET sqlc.narg('results_offset');
