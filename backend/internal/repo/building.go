package repo

import (
	"context"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
)

type BuildingRepo struct {
	*PostgresRepo
}

func NewBuildingRepo(repo *PostgresRepo) *BuildingRepo {
	return &BuildingRepo{repo}
}

func (r *BuildingRepo) GetByID(ctx context.Context, id int64) (*domain.Building, error) {
	ctx = queryNameToContext(ctx, "BuildingRepo.GetByID")
	row, err := r.queries.GetBuildingByID(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return buildingFromDB(row), nil
}

func (r *BuildingRepo) List(ctx context.Context) ([]domain.Building, error) {
	ctx = queryNameToContext(ctx, "BuildingRepo.List")
	rows, err := r.queries.ListBuildings(ctx)
	if err != nil {
		return nil, err
	}
	return buildingsFromDB(rows), nil
}

func buildingFromDB(b db.Building) *domain.Building {
	res := domain.Building(b)
	return &res
}

func buildingsFromDB(rows []db.Building) []domain.Building {
	result := make([]domain.Building, len(rows))
	for i, b := range rows {
		result[i] = domain.Building(b)
	}
	return result
}
