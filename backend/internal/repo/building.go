package repo

import (
	"context"
	"fmt"

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

func (r *BuildingRepo) Create(ctx context.Context, building *domain.Building) error {
	ctx = queryNameToContext(ctx, "BuildingRepo.Create")
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := db.New(tx)

	row, err := queries.CreateBuilding(ctx, db.CreateBuildingParams{
		Name:           building.Name,
		Address:        building.Address,
		ApartmentCount: building.ApartmentCount,
	})
	if err != nil {
		return err
	}

	if err := upsertApartmentsForBuilding(ctx, tx, row.ID, row.ApartmentCount); err != nil {
		return fmt.Errorf("create apartments for building: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	building.ID = row.ID
	building.CreatedAt = row.CreatedAt
	building.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *BuildingRepo) UpdateApartmentCount(ctx context.Context, id int64, apartmentCount int32) (*domain.Building, error) {
	ctx = queryNameToContext(ctx, "BuildingRepo.UpdateApartmentCount")
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := db.New(tx)
	row, err := queries.UpdateBuildingApartmentCount(ctx, db.UpdateBuildingApartmentCountParams{
		ID:             id,
		ApartmentCount: apartmentCount,
	})
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := upsertApartmentsForBuilding(ctx, tx, row.ID, row.ApartmentCount); err != nil {
		return nil, fmt.Errorf("create apartments for building: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return buildingFromDB(row), nil
}

func upsertApartmentsForBuilding(ctx context.Context, tx pgx.Tx, buildingID int64, apartmentCount int32) error {
	if apartmentCount <= 0 {
		return nil
	}

	const syncApartmentIDSequenceQuery = `
		SELECT setval(
			pg_get_serial_sequence('apartments', 'id'),
			COALESCE((SELECT MAX(id) FROM apartments), 0) + 1,
			false
		)
	`
	if _, err := tx.Exec(ctx, syncApartmentIDSequenceQuery); err != nil {
		return err
	}

	const upsertApartmentsQuery = `
		INSERT INTO apartments (building_id, number)
		SELECT $1, gs::text
		FROM generate_series(1, $2) AS gs
		ON CONFLICT DO NOTHING
	`

	_, err := tx.Exec(ctx, upsertApartmentsQuery, buildingID, apartmentCount)
	return err
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
