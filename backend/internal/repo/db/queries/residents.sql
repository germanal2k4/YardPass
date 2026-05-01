-- name: GetResidentByID :one
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, status, created_at, updated_at
FROM residents
WHERE id = $1;

-- name: GetResidentByTelegramID :one
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, status, created_at, updated_at
FROM residents
WHERE telegram_id = $1;

-- name: CreateResident :one
INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, car_plate, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;

-- name: UpdateResident :one
UPDATE residents
SET apartment_id = $2, telegram_id = $3, chat_id = $4, name = $5, phone = $6, car_plate = $7, status = $8
WHERE id = $1
RETURNING updated_at;

-- name: SetResidentCarPlate :exec
UPDATE residents
SET car_plate = $2
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
SELECT id, apartment_id, telegram_id, chat_id, name, phone, car_plate, status, created_at, updated_at
FROM residents
WHERE (sqlc.narg('filter_apartment_id')::bigint IS NULL OR apartment_id = sqlc.narg('filter_apartment_id'))
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR apartment_id IN (SELECT id FROM apartments WHERE building_id = sqlc.narg('filter_building_id')))
  AND (sqlc.narg('filter_status')::varchar IS NULL OR status = sqlc.narg('filter_status'))
ORDER BY created_at DESC
LIMIT sqlc.narg('max_results')
OFFSET sqlc.narg('results_offset');
