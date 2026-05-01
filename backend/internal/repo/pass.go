package repo

import (
	"context"
	"strings"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PassRepo struct {
	*PostgresRepo
}

func NewPassRepo(repo *PostgresRepo) *PassRepo {
	return &PassRepo{repo}
}

func (r *PassRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetByID")
	row, err := r.queries.GetPassByID(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return passFromRow(row), nil
}

func (r *PassRepo) GetByApartmentID(ctx context.Context, apartmentID int64, status string) ([]domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetByApartmentID")
	rows, err := r.queries.GetPassesByApartmentIDAndStatus(ctx, db.GetPassesByApartmentIDAndStatusParams{
		ApartmentID: apartmentID,
		Status:      status,
	})
	if err != nil {
		return nil, err
	}
	return passesFromRows(rows), nil
}

func (r *PassRepo) GetActiveByApartmentID(ctx context.Context, apartmentID int64) ([]domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetActiveByApartmentID")
	now := time.Now()
	rows, err := r.queries.GetActivePassesByApartmentID(ctx, db.GetActivePassesByApartmentIDParams{
		ApartmentID: apartmentID,
		ValidFrom:   now,
	})
	if err != nil {
		return nil, err
	}
	return passesFromRows(rows), nil
}

func (r *PassRepo) GetActiveByResidentID(ctx context.Context, residentID int64) ([]domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetActiveByResidentID")
	now := time.Now()
	rows, err := r.queries.GetActivePassesByResidentID(ctx, db.GetActivePassesByResidentIDParams{
		ResidentID: &residentID,
		ValidFrom:  now,
	})
	if err != nil {
		return nil, err
	}
	return passesFromRows(rows), nil
}

func (r *PassRepo) GetActiveByBuildingID(ctx context.Context, buildingID int64) ([]domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetActiveByBuildingID")
	now := time.Now()
	rows, err := r.queries.GetActivePassesByBuildingID(ctx, db.GetActivePassesByBuildingIDParams{
		BuildingID: buildingID,
		ValidFrom:  now,
	})
	if err != nil {
		return nil, err
	}
	return passesFromRows(rows), nil
}

func (r *PassRepo) GetActiveByCarPlate(ctx context.Context, normalizedCarPlate string, buildingID *int64) (*domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.GetActiveByCarPlate")
	now := time.Now()
	row, err := r.queries.GetActivePassByCarPlate(ctx, db.GetActivePassByCarPlateParams{
		CarPlate:         &normalizedCarPlate,
		ValidFrom:        now,
		FilterBuildingID: buildingID,
	})
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return passFromRow(row), nil
}

func (r *PassRepo) SearchByCarPlate(ctx context.Context, carPlate string, buildingID *int64, limit int) ([]domain.Pass, error) {
	ctx = queryNameToContext(ctx, "PassRepo.SearchByCarPlate")
	pattern := "%" + strings.ToUpper(strings.ReplaceAll(carPlate, " ", "")) + "%"
	rows, err := r.queries.SearchPassesByCarPlate(ctx, db.SearchPassesByCarPlateParams{
		CarPlatePattern:  pattern,
		MaxResults:       int32(limit),
		FilterBuildingID: buildingID,
	})
	if err != nil {
		return nil, err
	}
	return passesFromRows(rows), nil
}

func (r *PassRepo) CountActiveTodayByApartmentID(ctx context.Context, apartmentID int64) (int, error) {
	ctx = queryNameToContext(ctx, "PassRepo.CountActiveTodayByApartmentID")
	// Use local day boundary in Europe/Moscow to count daily limits
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		location = time.UTC
	}
	nowLocal := time.Now().In(location)
	todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, location)
	today := todayLocal.UTC()
	count, err := r.queries.CountActiveTodayByApartmentID(ctx, db.CountActiveTodayByApartmentIDParams{
		ApartmentID: apartmentID,
		CreatedAt:   today,
	})
	return int(count), err
}

func (r *PassRepo) CountActiveTodayByResidentID(ctx context.Context, residentID int64) (int, error) {
	ctx = queryNameToContext(ctx, "PassRepo.CountActiveTodayByResidentID")
	// Use local day boundary in Europe/Moscow to count daily limits
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		location = time.UTC
	}
	nowLocal := time.Now().In(location)
	todayLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, location)
	today := todayLocal.UTC()
	count, err := r.queries.CountActiveTodayByResidentID(ctx, db.CountActiveTodayByResidentIDParams{
		ResidentID: &residentID,
		CreatedAt:  today,
	})
	return int(count), err
}

func (r *PassRepo) Create(ctx context.Context, pass *domain.Pass) error {
	ctx = queryNameToContext(ctx, "PassRepo.Create")
	row, err := r.queries.CreatePass(ctx, db.CreatePassParams{
		ID:          pass.ID,
		ApartmentID: pass.ApartmentID,
		ResidentID:  pass.ResidentID,
		CarPlate:    pass.CarPlate,
		GuestName:   pass.GuestName,
		ValidFrom:   pass.ValidFrom,
		ValidTo:     pass.ValidTo,
		Status:      pass.Status,
	})
	if err != nil {
		return err
	}
	pass.CreatedAt = row.CreatedAt
	pass.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *PassRepo) Update(ctx context.Context, pass *domain.Pass) error {
	ctx = queryNameToContext(ctx, "PassRepo.Update")
	updatedAt, err := r.queries.UpdatePass(ctx, db.UpdatePassParams{
		ID:          pass.ID,
		ApartmentID: pass.ApartmentID,
		CarPlate:    pass.CarPlate,
		GuestName:   pass.GuestName,
		ValidFrom:   pass.ValidFrom,
		ValidTo:     pass.ValidTo,
		Status:      pass.Status,
	})
	if err != nil {
		return err
	}
	pass.UpdatedAt = updatedAt
	return nil
}

func (r *PassRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	ctx = queryNameToContext(ctx, "PassRepo.Revoke")
	return r.queries.RevokePass(ctx, id)
}

type passRow interface {
	db.GetPassByIDRow |
		db.GetActivePassByCarPlateRow |
		db.GetActivePassesByApartmentIDRow |
		db.GetActivePassesByResidentIDRow |
		db.GetActivePassesByBuildingIDRow |
		db.SearchPassesByCarPlateRow |
		db.GetPassesByApartmentIDAndStatusRow
}

func passFromRow[T passRow](r T) *domain.Pass {
	res := domain.Pass(r)
	return &res
}

func passesFromRows[T passRow](rows []T) []domain.Pass {
	result := make([]domain.Pass, len(rows))
	for i, r := range rows {
		result[i] = domain.Pass(r)
	}
	return result
}
