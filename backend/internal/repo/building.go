package repo

import (
	"context"

	"yardpass/internal/domain"

	"github.com/jackc/pgx/v5"
)

type BuildingRepo struct {
	*PostgresRepo
}

func NewBuildingRepo(repo *PostgresRepo) *BuildingRepo {
	return &BuildingRepo{repo}
}

func (r *BuildingRepo) GetByID(ctx context.Context, id int64) (*domain.Building, error) {
	query := `
		SELECT id, name, address, created_at, updated_at
		FROM buildings
		WHERE id = $1
	`

	var building domain.Building
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&building.ID,
		&building.Name,
		&building.Address,
		&building.CreatedAt,
		&building.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &building, nil
}

func (r *BuildingRepo) List(ctx context.Context) ([]*domain.Building, error) {
	query := `
		SELECT id, name, address, created_at, updated_at
		FROM buildings
		ORDER BY name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []*domain.Building
	for rows.Next() {
		var building domain.Building
		if err := rows.Scan(
			&building.ID,
			&building.Name,
			&building.Address,
			&building.CreatedAt,
			&building.UpdatedAt,
		); err != nil {
			return nil, err
		}
		buildings = append(buildings, &building)
	}

	return buildings, rows.Err()
}
