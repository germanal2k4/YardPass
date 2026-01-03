package repo

import (
	"context"

	"yardpass/internal/domain"
	"github.com/jackc/pgx/v5"
)

type ApartmentRepo struct {
	*PostgresRepo
}

func NewApartmentRepo(repo *PostgresRepo) *ApartmentRepo {
	return &ApartmentRepo{repo}
}

func (r *ApartmentRepo) GetByID(ctx context.Context, id int64) (*domain.Apartment, error) {
	query := `
		SELECT id, building_id, number, floor, created_at, updated_at
		FROM apartments
		WHERE id = $1
	`

	var apartment domain.Apartment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&apartment.ID,
		&apartment.BuildingID,
		&apartment.Number,
		&apartment.Floor,
		&apartment.CreatedAt,
		&apartment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &apartment, nil
}

func (r *ApartmentRepo) GetByBuildingID(ctx context.Context, buildingID int64) ([]*domain.Apartment, error) {
	query := `
		SELECT id, building_id, number, floor, created_at, updated_at
		FROM apartments
		WHERE building_id = $1
		ORDER BY number
	`

	rows, err := r.pool.Query(ctx, query, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apartments []*domain.Apartment
	for rows.Next() {
		var apartment domain.Apartment
		if err := rows.Scan(
			&apartment.ID,
			&apartment.BuildingID,
			&apartment.Number,
			&apartment.Floor,
			&apartment.CreatedAt,
			&apartment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		apartments = append(apartments, &apartment)
	}

	return apartments, rows.Err()
}

func (r *ApartmentRepo) GetByResidentTelegramID(ctx context.Context, telegramID int64) (*domain.Apartment, error) {
	query := `
		SELECT a.id, a.building_id, a.number, a.floor, a.created_at, a.updated_at
		FROM apartments a
		INNER JOIN residents r ON a.id = r.apartment_id
		WHERE r.telegram_id = $1
	`

	var apartment domain.Apartment
	err := r.pool.QueryRow(ctx, query, telegramID).Scan(
		&apartment.ID,
		&apartment.BuildingID,
		&apartment.Number,
		&apartment.Floor,
		&apartment.CreatedAt,
		&apartment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &apartment, nil
}

