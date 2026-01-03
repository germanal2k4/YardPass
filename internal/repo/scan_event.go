package repo

import (
	"context"
	"fmt"
	"time"

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
