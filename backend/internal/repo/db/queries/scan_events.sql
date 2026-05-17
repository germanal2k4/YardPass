-- name: CreateScanEvent :one
INSERT INTO scan_events (pass_id, guard_user_id, scanned_at, result, reason, meta)
VALUES ($1, $2, $3, $4, $5, sqlc.arg(meta)::jsonb)
RETURNING id;

-- name: ListScanEvents :many
SELECT id, pass_id, guard_user_id, scanned_at, result, reason, meta
FROM scan_events
WHERE (sqlc.narg('filter_pass_id')::uuid IS NULL OR pass_id = sqlc.narg('filter_pass_id'))
  AND (sqlc.narg('filter_guard_user_id')::bigint IS NULL OR guard_user_id = sqlc.narg('filter_guard_user_id'))
  AND (sqlc.narg('filter_result')::varchar IS NULL OR result = sqlc.narg('filter_result'))
  AND (sqlc.narg('filter_from')::timestamp IS NULL OR scanned_at >= sqlc.narg('filter_from'))
  AND (sqlc.narg('filter_to')::timestamp IS NULL OR scanned_at <= sqlc.narg('filter_to'))
ORDER BY scanned_at DESC
LIMIT sqlc.narg('max_results')
OFFSET sqlc.narg('results_offset');

-- name: CountValidScansToday :one
SELECT COUNT(*)
FROM scan_events
WHERE result = 'valid' AND scanned_at >= $1;

-- name: GetScanEventStatistics :one
SELECT
    COUNT(*) as total_scans,
    COUNT(*) FILTER (WHERE se.result = 'valid') as valid_scans,
    COUNT(*) FILTER (WHERE se.result = 'invalid') as invalid_scans,
    COUNT(DISTINCT se.pass_id) as unique_passes,
    COUNT(DISTINCT se.guard_user_id) as unique_guards
FROM scan_events se
INNER JOIN passes p ON se.pass_id = p.id
INNER JOIN apartments a ON p.apartment_id = a.id
WHERE (sqlc.narg('filter_from')::timestamp IS NULL OR se.scanned_at >= sqlc.narg('filter_from'))
  AND (sqlc.narg('filter_to')::timestamp IS NULL OR se.scanned_at <= sqlc.narg('filter_to'))
  AND (sqlc.narg('filter_building_id')::bigint IS NULL OR a.building_id = sqlc.narg('filter_building_id'));

-- name: GetScanEventsWithDetails :many
SELECT
    se.id, se.pass_id, se.guard_user_id, se.scanned_at, se.result, se.reason, se.meta,
    p.car_plate, p.guest_name, a.number as apartment_number, a.building_id,
    b.name as building_name,
    u.username as guard_username
FROM scan_events se
INNER JOIN passes p ON se.pass_id = p.id
INNER JOIN apartments a ON p.apartment_id = a.id
INNER JOIN buildings b ON a.building_id = b.id
LEFT JOIN users u ON se.guard_user_id = u.id
WHERE (sqlc.narg('filter_building_id')::bigint IS NULL OR a.building_id = sqlc.narg('filter_building_id'))
  AND (sqlc.narg('filter_pass_id')::uuid IS NULL OR se.pass_id = sqlc.narg('filter_pass_id'))
  AND (sqlc.narg('filter_guard_user_id')::bigint IS NULL OR se.guard_user_id = sqlc.narg('filter_guard_user_id'))
  AND (sqlc.narg('filter_result')::varchar IS NULL OR se.result = sqlc.narg('filter_result'))
  AND (sqlc.narg('filter_apartment_number')::varchar IS NULL OR a.number ILIKE '%' || sqlc.narg('filter_apartment_number') || '%')
  AND (sqlc.narg('filter_car_plate')::varchar IS NULL OR p.car_plate ILIKE '%' || sqlc.narg('filter_car_plate') || '%')
  AND (sqlc.narg('filter_from')::timestamp IS NULL OR se.scanned_at >= sqlc.narg('filter_from'))
  AND (sqlc.narg('filter_to')::timestamp IS NULL OR se.scanned_at <= sqlc.narg('filter_to'))
ORDER BY se.scanned_at DESC
LIMIT sqlc.narg('max_results')
OFFSET sqlc.narg('results_offset');
