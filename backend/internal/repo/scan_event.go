package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"yardpass/internal/domain"
)

type ScanEventRepo struct {
	*PostgresRepo
}

func NewScanEventRepo(repo *PostgresRepo) *ScanEventRepo {
	return &ScanEventRepo{repo}
}

func (r *ScanEventRepo) Create(ctx context.Context, event *domain.ScanEvent) error {
	query := `
		INSERT INTO scan_events (pass_id, guard_user_id, scanned_at, result, reason, meta)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		event.PassID,
		event.GuardUserID,
		event.ScannedAt,
		event.Result,
		event.Reason,
		event.Meta,
	).Scan(&event.ID)

	return err
}

func (r *ScanEventRepo) List(ctx context.Context, filters domain.ScanEventFilters) ([]*domain.ScanEvent, error) {
	query := `
		SELECT id, pass_id, guard_user_id, scanned_at, result, reason, meta
		FROM scan_events
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filters.PassID != nil {
		query += fmt.Sprintf(` AND pass_id = $%d`, argPos)
		args = append(args, *filters.PassID)
		argPos++
	}

	if filters.GuardUserID != nil {
		query += fmt.Sprintf(` AND guard_user_id = $%d`, argPos)
		args = append(args, *filters.GuardUserID)
		argPos++
	}

	if filters.Result != nil {
		query += fmt.Sprintf(` AND result = $%d`, argPos)
		args = append(args, *filters.Result)
		argPos++
	}

	if filters.From != nil {
		query += fmt.Sprintf(` AND scanned_at >= $%d`, argPos)
		args = append(args, *filters.From)
		argPos++
	}

	if filters.To != nil {
		query += fmt.Sprintf(` AND scanned_at <= $%d`, argPos)
		args = append(args, *filters.To)
		argPos++
	}

	query += ` ORDER BY scanned_at DESC`

	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argPos)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.ScanEvent
	for rows.Next() {
		var event domain.ScanEvent
		if err := rows.Scan(
			&event.ID,
			&event.PassID,
			&event.GuardUserID,
			&event.ScannedAt,
			&event.Result,
			&event.Reason,
			&event.Meta,
		); err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, rows.Err()
}

func (r *ScanEventRepo) CountValidScansToday(ctx context.Context) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	query := `
		SELECT COUNT(*)
		FROM scan_events
		WHERE result = 'valid' AND scanned_at >= $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, today).Scan(&count)
	return count, err
}

func (r *ScanEventRepo) GetStatistics(ctx context.Context, from, to *time.Time, buildingID *int64) (*Statistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_scans,
			COUNT(*) FILTER (WHERE result = 'valid') as valid_scans,
			COUNT(*) FILTER (WHERE result = 'invalid') as invalid_scans,
			COUNT(DISTINCT pass_id) as unique_passes,
			COUNT(DISTINCT guard_user_id) as unique_guards
		FROM scan_events se
		INNER JOIN passes p ON se.pass_id = p.id
		INNER JOIN apartments a ON p.apartment_id = a.id
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if from != nil {
		query += fmt.Sprintf(` AND se.scanned_at >= $%d`, argPos)
		args = append(args, *from)
		argPos++
	}

	if to != nil {
		query += fmt.Sprintf(` AND se.scanned_at <= $%d`, argPos)
		args = append(args, *to)
		argPos++
	}

	if buildingID != nil {
		query += fmt.Sprintf(` AND a.building_id = $%d`, argPos)
		args = append(args, *buildingID)
		argPos++
	}

	var stats Statistics
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalScans,
		&stats.ValidScans,
		&stats.InvalidScans,
		&stats.UniquePasses,
		&stats.UniqueGuards,
	)

	return &stats, err
}

type Statistics struct {
	TotalScans   int
	ValidScans   int
	InvalidScans int
	UniquePasses int
	UniqueGuards int
}

func (r *ScanEventRepo) GetEventsWithDetails(ctx context.Context, filters domain.ScanEventFilters, buildingID *int64) ([]*ScanEventWithDetails, error) {
	query := `
		SELECT 
			se.id, se.pass_id, se.guard_user_id, se.scanned_at, se.result, se.reason, se.meta,
			p.car_plate, a.number as apartment_number, a.building_id,
			u.username as guard_username
		FROM scan_events se
		INNER JOIN passes p ON se.pass_id = p.id
		INNER JOIN apartments a ON p.apartment_id = a.id
		LEFT JOIN users u ON se.guard_user_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if buildingID != nil {
		query += fmt.Sprintf(` AND a.building_id = $%d`, argPos)
		args = append(args, *buildingID)
		argPos++
	}

	if filters.PassID != nil {
		query += fmt.Sprintf(` AND se.pass_id = $%d`, argPos)
		args = append(args, *filters.PassID)
		argPos++
	}

	if filters.GuardUserID != nil {
		query += fmt.Sprintf(` AND se.guard_user_id = $%d`, argPos)
		args = append(args, *filters.GuardUserID)
		argPos++
	}

	if filters.Result != nil {
		query += fmt.Sprintf(` AND se.result = $%d`, argPos)
		args = append(args, *filters.Result)
		argPos++
	}

	if filters.From != nil {
		query += fmt.Sprintf(` AND se.scanned_at >= $%d`, argPos)
		args = append(args, *filters.From)
		argPos++
	}

	if filters.To != nil {
		query += fmt.Sprintf(` AND se.scanned_at <= $%d`, argPos)
		args = append(args, *filters.To)
		argPos++
	}

	query += ` ORDER BY se.scanned_at DESC`

	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argPos)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*ScanEventWithDetails
	for rows.Next() {
		var event ScanEventWithDetails
		var guardUsername *string
		if err := rows.Scan(
			&event.ID,
			&event.PassID,
			&event.GuardUserID,
			&event.ScannedAt,
			&event.Result,
			&event.Reason,
			&event.Meta,
			&event.CarPlate,
			&event.ApartmentNumber,
			&event.BuildingID,
			&guardUsername,
		); err != nil {
			return nil, err
		}
		if guardUsername != nil {
			event.GuardUsername = *guardUsername
		}
		events = append(events, &event)
	}

	return events, rows.Err()
}

type ScanEventWithDetails struct {
	ID              int64
	PassID          uuid.UUID
	GuardUserID     int64
	GuardUsername   string
	ScannedAt       time.Time
	Result          string
	Reason          *string
	Meta            *string
	CarPlate        string
	ApartmentNumber string
	BuildingID      int64
}
