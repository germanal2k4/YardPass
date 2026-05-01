package repo

import (
	"context"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
)

type ApartmentRepo struct {
	*PostgresRepo
}

func NewApartmentRepo(repo *PostgresRepo) *ApartmentRepo {
	return &ApartmentRepo{repo}
}

func (r *ApartmentRepo) GetByID(ctx context.Context, id int64) (*domain.Apartment, error) {
	ctx = queryNameToContext(ctx, "ApartmentRepo.GetByID")
	row, err := r.queries.GetApartmentByID(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return apartmentFromDB(row), nil
}

func (r *ApartmentRepo) GetByBuildingID(ctx context.Context, buildingID int64) ([]domain.Apartment, error) {
	ctx = queryNameToContext(ctx, "ApartmentRepo.GetByBuildingID")
	rows, err := r.queries.GetApartmentsByBuildingID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	return apartmentsFromDB(rows), nil
}

func (r *ApartmentRepo) GetByResidentTelegramID(ctx context.Context, telegramID int64) (*domain.Apartment, error) {
	ctx = queryNameToContext(ctx, "ApartmentRepo.GetByResidentTelegramID")
	row, err := r.queries.GetApartmentByResidentTelegramID(ctx, telegramID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return apartmentFromDB(row), nil
}

func apartmentFromDB(a db.Apartment) *domain.Apartment {
	res := domain.Apartment(a)
	return &res
}

func apartmentsFromDB(rows []db.Apartment) []domain.Apartment {
	result := make([]domain.Apartment, len(rows))
	for i, a := range rows {
		result[i] = domain.Apartment(a)
	}
	return result
}
