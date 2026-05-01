-- name: GetRuleByBuildingID :one
SELECT id, building_id, quiet_hours_start, quiet_hours_end,
       daily_pass_limit_per_apartment, max_pass_duration_hours, created_at, updated_at
FROM rules
WHERE building_id = $1;

-- name: CreateRule :one
INSERT INTO rules (building_id, quiet_hours_start, quiet_hours_end,
                   daily_pass_limit_per_apartment, max_pass_duration_hours)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at;

-- name: UpdateRule :one
UPDATE rules
SET quiet_hours_start = $2, quiet_hours_end = $3,
    daily_pass_limit_per_apartment = $4, max_pass_duration_hours = $5
WHERE id = $1
RETURNING updated_at;
