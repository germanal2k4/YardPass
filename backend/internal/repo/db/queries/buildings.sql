-- name: GetBuildingByID :one
SELECT id, name, address, apartment_count, created_at, updated_at
FROM buildings
WHERE id = $1;

-- name: ListBuildings :many
SELECT id, name, address, apartment_count, created_at, updated_at
FROM buildings
ORDER BY name;

-- name: CreateBuilding :one
INSERT INTO buildings (name, address, apartment_count)
VALUES ($1, $2, $3)
RETURNING id, name, address, apartment_count, created_at, updated_at;

-- name: UpdateBuildingApartmentCount :one
UPDATE buildings
SET apartment_count = $2
WHERE id = $1
RETURNING id, name, address, apartment_count, created_at, updated_at;
