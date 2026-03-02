-- name: GetPassByID :one
SELECT id, apartment_id, resident_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
FROM passes
WHERE id = $1;

-- name: GetPassesByApartmentIDAndStatus :many
SELECT id, apartment_id, resident_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
FROM passes
WHERE apartment_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: GetActivePassesByApartmentID :many
SELECT id, apartment_id, resident_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
FROM passes
WHERE apartment_id = $1
  AND status = 'active'
  AND valid_from <= $2
  AND valid_to >= $2
ORDER BY created_at DESC;

-- name: GetActivePassesByResidentID :many
SELECT id, apartment_id, resident_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
FROM passes
WHERE resident_id = $1
  AND status = 'active'
  AND valid_from <= $2
  AND valid_to >= $2
ORDER BY created_at DESC;

-- name: GetActivePassesByBuildingID :many
SELECT p.id, p.apartment_id, p.resident_id, p.car_plate, p.guest_name, p.valid_from, p.valid_to, p.status, p.created_at, p.updated_at
FROM passes p
INNER JOIN apartments a ON p.apartment_id = a.id
WHERE a.building_id = $1
  AND p.status = 'active'
  AND p.valid_from <= $2
  AND p.valid_to >= $2
ORDER BY p.created_at DESC;

-- name: GetActivePassByCarPlate :one
SELECT p.id, p.apartment_id, p.resident_id, p.car_plate, p.guest_name, p.valid_from, p.valid_to, p.status, p.created_at, p.updated_at
FROM passes p
INNER JOIN apartments a ON p.apartment_id = a.id
WHERE p.car_plate = $1
  AND p.status = 'active'
  AND p.valid_from <= $2
  AND p.valid_to >= $2
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR a.building_id = sqlc.narg('filter_building_id'))
ORDER BY p.created_at DESC
LIMIT 1;

-- name: SearchPassesByCarPlate :many
SELECT p.id, p.apartment_id, p.resident_id, p.car_plate, p.guest_name, p.valid_from, p.valid_to, p.status, p.created_at, p.updated_at
FROM passes p
INNER JOIN apartments a ON p.apartment_id = a.id
WHERE UPPER(REPLACE(p.car_plate, ' ', '')) LIKE sqlc.arg('car_plate_pattern')::varchar
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR a.building_id = sqlc.narg('filter_building_id'))
ORDER BY p.created_at DESC
LIMIT sqlc.arg('max_results')::int;

-- name: CountActiveTodayByApartmentID :one
SELECT COUNT(*)
FROM passes
WHERE apartment_id = $1
  AND status = 'active'
  AND created_at >= $2;

-- name: CountActiveTodayByResidentID :one
SELECT COUNT(*)
FROM passes
WHERE resident_id = $1
  AND status = 'active'
  AND created_at >= $2;

-- name: CreatePass :one
INSERT INTO passes (id, apartment_id, resident_id, car_plate, guest_name, valid_from, valid_to, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING created_at, updated_at;

-- name: UpdatePass :one
UPDATE passes
SET apartment_id = $2, car_plate = $3, guest_name = $4, valid_from = $5, valid_to = $6, status = $7
WHERE id = $1
RETURNING updated_at;

-- name: RevokePass :exec
UPDATE passes
SET status = 'revoked'
WHERE id = $1;
