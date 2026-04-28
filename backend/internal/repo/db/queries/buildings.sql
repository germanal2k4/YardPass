-- name: GetBuildingByID :one
SELECT id, name, address, created_at, updated_at
FROM buildings
WHERE id = $1;

-- name: ListBuildings :many
SELECT id, name, address, created_at, updated_at
FROM buildings
ORDER BY name;

-- name: CreateBuilding :one
INSERT INTO buildings (name, address)
VALUES ($1, $2)
RETURNING id, name, address, created_at, updated_at;
