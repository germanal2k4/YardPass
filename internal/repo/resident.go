package repo

import (
	"context"

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

