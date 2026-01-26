package repo

import (
	"context"
	"fmt"

	"yardpass/internal/domain"

	"github.com/jackc/pgx/v5"
)

type ResidentRepo struct {
	*PostgresRepo
}

func NewResidentRepo(repo *PostgresRepo) *ResidentRepo {
	return &ResidentRepo{repo}
}

func (r *ResidentRepo) GetByID(ctx context.Context, id int64) (*domain.Resident, error) {
	query := `
		SELECT id, apartment_id, telegram_id, chat_id, name, phone, status, created_at, updated_at
		FROM residents
		WHERE id = $1
	`

	var resident domain.Resident
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&resident.ID,
		&resident.ApartmentID,
		&resident.TelegramID,
		&resident.ChatID,
		&resident.Name,
		&resident.Phone,
		&resident.Status,
		&resident.CreatedAt,
		&resident.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &resident, nil
}

func (r *ResidentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Resident, error) {
	query := `
		SELECT id, apartment_id, telegram_id, chat_id, name, phone, status, created_at, updated_at
		FROM residents
		WHERE telegram_id = $1
	`

	var resident domain.Resident
	err := r.pool.QueryRow(ctx, query, telegramID).Scan(
		&resident.ID,
		&resident.ApartmentID,
		&resident.TelegramID,
		&resident.ChatID,
		&resident.Name,
		&resident.Phone,
		&resident.Status,
		&resident.CreatedAt,
		&resident.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &resident, nil
}

func (r *ResidentRepo) Create(ctx context.Context, resident *domain.Resident) error {
	query := `
		INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		resident.ApartmentID,
		resident.TelegramID,
		resident.ChatID,
		resident.Name,
		resident.Phone,
		resident.Status,
	).Scan(&resident.ID, &resident.CreatedAt, &resident.UpdatedAt)

	return err
}

func (r *ResidentRepo) Update(ctx context.Context, resident *domain.Resident) error {
	query := `
		UPDATE residents
		SET apartment_id = $2, telegram_id = $3, chat_id = $4, name = $5, phone = $6, status = $7
		WHERE id = $1
		RETURNING updated_at
	`

	return r.pool.QueryRow(ctx, query,
		resident.ID,
		resident.ApartmentID,
		resident.TelegramID,
		resident.ChatID,
		resident.Name,
		resident.Phone,
		resident.Status,
	).Scan(&resident.UpdatedAt)
}

func (r *ResidentRepo) BulkCreate(ctx context.Context, residents []*domain.Resident) error {
	if len(residents) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO residents (apartment_id, telegram_id, chat_id, name, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (telegram_id) DO UPDATE SET
			apartment_id = EXCLUDED.apartment_id,
			chat_id = EXCLUDED.chat_id,
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			status = EXCLUDED.status
		RETURNING id, created_at, updated_at
	`

	for _, resident := range residents {
		err := tx.QueryRow(ctx, query,
			resident.ApartmentID,
			resident.TelegramID,
			resident.ChatID,
			resident.Name,
			resident.Phone,
			resident.Status,
		).Scan(&resident.ID, &resident.CreatedAt, &resident.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ResidentRepo) List(ctx context.Context, filters domain.ResidentFilters) ([]*domain.Resident, error) {
	query := `
		SELECT id, apartment_id, telegram_id, chat_id, name, phone, status, created_at, updated_at
		FROM residents
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filters.ApartmentID != nil {
		query += fmt.Sprintf(` AND apartment_id = $%d`, argPos)
		args = append(args, *filters.ApartmentID)
		argPos++
	}

	if filters.BuildingID != nil {
		query += ` AND apartment_id IN (SELECT id FROM apartments WHERE building_id = $` + fmt.Sprintf("%d", argPos) + `)`
		args = append(args, *filters.BuildingID)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

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

	var residents []*domain.Resident
	for rows.Next() {
		var resident domain.Resident
		if err := rows.Scan(
			&resident.ID,
			&resident.ApartmentID,
			&resident.TelegramID,
			&resident.ChatID,
			&resident.Name,
			&resident.Phone,
			&resident.Status,
			&resident.CreatedAt,
			&resident.UpdatedAt,
		); err != nil {
			return nil, err
		}
		residents = append(residents, &resident)
	}

	return residents, rows.Err()
}

func (r *ResidentRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM residents WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
