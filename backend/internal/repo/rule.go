package repo

import (
	"context"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
)

type RuleRepo struct {
	*PostgresRepo
}

func NewRuleRepo(repo *PostgresRepo) *RuleRepo {
	return &RuleRepo{repo}
}

func (r *RuleRepo) GetByBuildingID(ctx context.Context, buildingID int64) (*domain.Rule, error) {
	ctx = queryNameToContext(ctx, "RuleRepo.GetByBuildingID")
	row, err := r.queries.GetRuleByBuildingID(ctx, buildingID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ruleFromDB(row), nil
}

func (r *RuleRepo) Create(ctx context.Context, rule *domain.Rule) error {
	ctx = queryNameToContext(ctx, "RuleRepo.Create")
	row, err := r.queries.CreateRule(ctx, db.CreateRuleParams{
		BuildingID:                 rule.BuildingID,
		QuietHoursStart:            rule.QuietHoursStart,
		QuietHoursEnd:              rule.QuietHoursEnd,
		DailyPassLimitPerApartment: int32(rule.DailyPassLimitPerApartment),
		MaxPassDurationHours:       int32(rule.MaxPassDurationHours),
	})
	if err != nil {
		return err
	}
	rule.ID = row.ID
	rule.CreatedAt = row.CreatedAt
	rule.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *RuleRepo) Update(ctx context.Context, rule *domain.Rule) error {
	ctx = queryNameToContext(ctx, "RuleRepo.Update")
	updatedAt, err := r.queries.UpdateRule(ctx, db.UpdateRuleParams{
		ID:                         rule.ID,
		QuietHoursStart:            rule.QuietHoursStart,
		QuietHoursEnd:              rule.QuietHoursEnd,
		DailyPassLimitPerApartment: int32(rule.DailyPassLimitPerApartment),
		MaxPassDurationHours:       int32(rule.MaxPassDurationHours),
	})
	if err != nil {
		return err
	}
	rule.UpdatedAt = updatedAt
	return nil
}

func ruleFromDB(r db.Rule) *domain.Rule {
	res := domain.Rule(r)
	return &res
}
