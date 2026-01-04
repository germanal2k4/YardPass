package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"yardpass/internal/domain"
	"github.com/jackc/pgx/v5"
)

type PassRepo struct {
	*PostgresRepo
}

func NewPassRepo(repo *PostgresRepo) *PassRepo {
	return &PassRepo{repo}
}

func (r *PassRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pass, error) {
	query := `
		SELECT id, apartment_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
		FROM passes
		WHERE id = $1
	`

	var pass domain.Pass
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pass.ID,
		&pass.ApartmentID,
		&pass.CarPlate,
		&pass.GuestName,
		&pass.ValidFrom,
		&pass.ValidTo,
		&pass.Status,
		&pass.CreatedAt,
		&pass.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &pass, nil
}

func (r *PassRepo) GetByApartmentID(ctx context.Context, apartmentID int64, status string) ([]*domain.Pass, error) {
	query := `
		SELECT id, apartment_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
		FROM passes
		WHERE apartment_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, apartmentID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passes []*domain.Pass
	for rows.Next() {
		var pass domain.Pass
		if err := rows.Scan(
			&pass.ID,
			&pass.ApartmentID,
			&pass.CarPlate,
			&pass.GuestName,
			&pass.ValidFrom,
			&pass.ValidTo,
			&pass.Status,
			&pass.CreatedAt,
			&pass.UpdatedAt,
		); err != nil {
			return nil, err
		}
		passes = append(passes, &pass)
	}

	return passes, rows.Err()
}

func (r *PassRepo) GetActiveByApartmentID(ctx context.Context, apartmentID int64) ([]*domain.Pass, error) {
	now := time.Now()
	query := `
		SELECT id, apartment_id, car_plate, guest_name, valid_from, valid_to, status, created_at, updated_at
		FROM passes
		WHERE apartment_id = $1 
			AND status = 'active'
			AND valid_from <= $2
			AND valid_to >= $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, apartmentID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passes []*domain.Pass
	for rows.Next() {
		var pass domain.Pass
		if err := rows.Scan(
			&pass.ID,
			&pass.ApartmentID,
			&pass.CarPlate,
			&pass.GuestName,
			&pass.ValidFrom,
			&pass.ValidTo,
			&pass.Status,
			&pass.CreatedAt,
			&pass.UpdatedAt,
		); err != nil {
			return nil, err
		}
		passes = append(passes, &pass)
	}

	return passes, rows.Err()
}

func (r *PassRepo) CountActiveTodayByApartmentID(ctx context.Context, apartmentID int64) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	query := `
		SELECT COUNT(*)
		FROM passes
		WHERE apartment_id = $1
			AND status = 'active'
			AND created_at >= $2
	`

	var count int
	err := r.pool.QueryRow(ctx, query, apartmentID, today).Scan(&count)
	return count, err
}

func (r *PassRepo) Create(ctx context.Context, pass *domain.Pass) error {
	query := `
		INSERT INTO passes (id, apartment_id, car_plate, guest_name, valid_from, valid_to, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		pass.ID,
		pass.ApartmentID,
		pass.CarPlate,
		pass.GuestName,
		pass.ValidFrom,
		pass.ValidTo,
		pass.Status,
	).Scan(&pass.CreatedAt, &pass.UpdatedAt)

	return err
}

func (r *PassRepo) Update(ctx context.Context, pass *domain.Pass) error {
	query := `
		UPDATE passes
		SET apartment_id = $2, car_plate = $3, guest_name = $4, valid_from = $5, valid_to = $6, status = $7
		WHERE id = $1
		RETURNING updated_at
	`

	return r.pool.QueryRow(ctx, query,
		pass.ID,
		pass.ApartmentID,
		pass.CarPlate,
		pass.GuestName,
		pass.ValidFrom,
		pass.ValidTo,
		pass.Status,
	).Scan(&pass.UpdatedAt)
}

func (r *PassRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE passes
		SET status = 'revoked'
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

