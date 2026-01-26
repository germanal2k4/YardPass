package repo

import (
	"context"

	"yardpass/internal/domain"

	"github.com/jackc/pgx/v5"
)

type RuleRepo struct {
	*PostgresRepo
}

func NewRuleRepo(repo *PostgresRepo) *RuleRepo {
	return &RuleRepo{repo}
}

func (r *RuleRepo) GetByBuildingID(ctx context.Context, buildingID int64) (*domain.Rule, error) {
	query := `
		SELECT id, building_id, quiet_hours_start, quiet_hours_end,
		       daily_pass_limit_per_apartment, max_pass_duration_hours, created_at, updated_at
		FROM rules
		WHERE building_id = $1
	`

	var rule domain.Rule
	err := r.pool.QueryRow(ctx, query, buildingID).Scan(
		&rule.ID,
		&rule.BuildingID,
		&rule.QuietHoursStart,
		&rule.QuietHoursEnd,
		&rule.DailyPassLimitPerApartment,
		&rule.MaxPassDurationHours,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *RuleRepo) Create(ctx context.Context, rule *domain.Rule) error {
	query := `
		INSERT INTO rules (building_id, quiet_hours_start, quiet_hours_end,
		                   daily_pass_limit_per_apartment, max_pass_duration_hours)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		rule.BuildingID,
		rule.QuietHoursStart,
		rule.QuietHoursEnd,
		rule.DailyPassLimitPerApartment,
		rule.MaxPassDurationHours,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	return err
}

func (r *RuleRepo) Update(ctx context.Context, rule *domain.Rule) error {
	query := `
		UPDATE rules
		SET quiet_hours_start = $2, quiet_hours_end = $3,
		    daily_pass_limit_per_apartment = $4, max_pass_duration_hours = $5
		WHERE id = $1
		RETURNING updated_at
	`

	return r.pool.QueryRow(ctx, query,
		rule.ID,
		rule.QuietHoursStart,
		rule.QuietHoursEnd,
		rule.DailyPassLimitPerApartment,
		rule.MaxPassDurationHours,
	).Scan(&rule.UpdatedAt)
}
