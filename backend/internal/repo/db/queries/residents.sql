-- name: GetResidentByID :one
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, timezone, status, created_at, updated_at
FROM residents
WHERE id = $1;

-- name: GetResidentByTelegramID :one
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, timezone, status, created_at, updated_at
FROM residents
WHERE telegram_id = $1;

-- name: ListResidentTelegramIDsIn :many
SELECT telegram_id
FROM residents
WHERE telegram_id = ANY($1::bigint[]);

-- name: CreateResident :one
INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, car_plate, timezone, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at;

-- name: UpdateResident :one
UPDATE residents
SET apartment_id = $2, telegram_id = $3, chat_id = $4, name = $5, phone = $6, car_plate = $7, timezone = $8, status = $9
WHERE id = $1
RETURNING updated_at;

-- name: SetResidentCarPlate :exec
UPDATE residents
SET car_plate = $2
WHERE id = $1;

-- name: SetResidentTimezone :exec
UPDATE residents
SET timezone = $2
WHERE id = $1;

-- name: UpsertResident :one
INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, car_plate, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (telegram_id) DO UPDATE SET
    apartment_id = EXCLUDED.apartment_id,
    chat_id = EXCLUDED.chat_id,
    name = EXCLUDED.name,
    phone = EXCLUDED.phone,
    car_plate = EXCLUDED.car_plate,
    status = EXCLUDED.status
RETURNING id, created_at, updated_at;

-- name: DeleteResident :exec
DELETE FROM residents WHERE id = $1;

-- name: ListResidents :many
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, timezone, status, created_at, updated_at
FROM residents
WHERE (sqlc.narg('filter_apartment_id')::bigint IS NULL OR apartment_id = sqlc.narg('filter_apartment_id'))
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR apartment_id IN (SELECT id FROM apartments WHERE building_id = sqlc.narg('filter_building_id')))
  AND (sqlc.narg('filter_status')::varchar IS NULL OR status = sqlc.narg('filter_status'))
ORDER BY created_at DESC
LIMIT sqlc.narg('max_results')
OFFSET sqlc.narg('results_offset');

-- name: ListActiveResidentsWithCarPlate :many
SELECT r.id, r.apartment_id, r.telegram_id, r.chat_id, r.name, r.phone, r.car_plate, r.timezone, r.status, r.created_at, r.updated_at
FROM residents r
INNER JOIN apartments a ON r.apartment_id = a.id
WHERE r.status = 'active'
  AND r.car_plate IS NOT NULL
  AND BTRIM(r.car_plate::text) <> ''
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR a.building_id = sqlc.narg('filter_building_id'))
ORDER BY r.id;
