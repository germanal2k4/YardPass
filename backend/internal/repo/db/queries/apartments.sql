-- name: GetApartmentByID :one
SELECT id, building_id, number, floor, created_at, updated_at
FROM apartments
WHERE id = $1;

-- name: GetApartmentsByBuildingID :many
SELECT id, building_id, number, floor, created_at, updated_at
FROM apartments
WHERE building_id = $1
ORDER BY number;

-- name: GetApartmentByResidentTelegramID :one
SELECT a.id, a.building_id, a.number, a.floor, a.created_at, a.updated_at
FROM apartments a
INNER JOIN residents r ON a.id = r.apartment_id
WHERE r.telegram_id = $1;
