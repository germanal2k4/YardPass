package repo

import (
	"context"

	"yardpass/internal/domain"
	"yardpass/internal/observability/logger"
	"yardpass/internal/repo/db"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ResidentRepo struct {
	*PostgresRepo
}

func NewResidentRepo(repo *PostgresRepo) *ResidentRepo {
	return &ResidentRepo{repo}
}

func (r *ResidentRepo) GetByID(ctx context.Context, id int64) (*domain.Resident, error) {
	ctx = queryNameToContext(ctx, "ResidentRepo.GetByID")
	row, err := r.queries.GetResidentByID(ctx, id)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return residentFromDB(row), nil
}

func (r *ResidentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Resident, error) {
	ctx = queryNameToContext(ctx, "ResidentRepo.GetByTelegramID")
	row, err := r.queries.GetResidentByTelegramID(ctx, telegramID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return residentFromDB(row), nil
}

func (r *ResidentRepo) Create(ctx context.Context, resident *domain.Resident) error {
	ctx = queryNameToContext(ctx, "ResidentRepo.Create")
	row, err := r.queries.CreateResident(ctx, db.CreateResidentParams{
		ApartmentID: resident.ApartmentID,
		TelegramID:  resident.TelegramID,
		ChatID:      resident.ChatID,
		Name:        resident.Name,
		Phone:       resident.Phone,
		Status:      resident.Status,
	})
	if err != nil {
		return err
	}
	resident.ID = row.ID
	resident.CreatedAt = row.CreatedAt
	resident.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ResidentRepo) Update(ctx context.Context, resident *domain.Resident) error {
	ctx = queryNameToContext(ctx, "ResidentRepo.Update")
	updatedAt, err := r.queries.UpdateResident(ctx, db.UpdateResidentParams{
		ID:          resident.ID,
		ApartmentID: resident.ApartmentID,
		TelegramID:  resident.TelegramID,
		ChatID:      resident.ChatID,
		Name:        resident.Name,
		Phone:       resident.Phone,
		Status:      resident.Status,
	})
	if err != nil {
		return err
	}
	resident.UpdatedAt = updatedAt
	return nil
}

func (r *ResidentRepo) BulkCreate(ctx context.Context, residents []domain.Resident) error {
	ctx = queryNameToContext(ctx, "ResidentRepo.BulkCreate")
	if len(residents) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			logger.FromContext(ctx).Error("Failed to rollback transaction", zap.Error(err))
		}
	}()

	qtx := r.queries.WithTx(tx)
	for _, resident := range residents {
		row, err := qtx.UpsertResident(ctx, db.UpsertResidentParams{
			ApartmentID: resident.ApartmentID,
			TelegramID:  resident.TelegramID,
			ChatID:      resident.ChatID,
			Name:        resident.Name,
			Phone:       resident.Phone,
			Status:      resident.Status,
		})
		if err != nil {
			return err
		}
		resident.ID = row.ID
		resident.CreatedAt = row.CreatedAt
		resident.UpdatedAt = row.UpdatedAt
	}

	return tx.Commit(ctx)
}

func (r *ResidentRepo) List(ctx context.Context, filters domain.ResidentFilters) ([]domain.Resident, error) {
	ctx = queryNameToContext(ctx, "ResidentRepo.List")
	rows, err := r.queries.ListResidents(ctx, db.ListResidentsParams{
		FilterApartmentID: filters.ApartmentID,
		FilterBuildingID:  filters.BuildingID,
		FilterStatus:      filters.Status,
		MaxResults:        intToInt32Ptr(filters.Limit),
		ResultsOffset:     intToInt32Ptr(filters.Offset),
	})
	if err != nil {
		return nil, err
	}
	return residentsFromDB(rows), nil
}

func (r *ResidentRepo) Delete(ctx context.Context, id int64) error {
	ctx = queryNameToContext(ctx, "ResidentRepo.Delete")
	return r.queries.DeleteResident(ctx, id)
}

func residentFromDB(row db.Resident) *domain.Resident {
	res := domain.Resident(row)
	return &res
}

func residentsFromDB(rows []db.Resident) []domain.Resident {
	result := make([]domain.Resident, len(rows))
	for i, row := range rows {
		result[i] = domain.Resident(row)
	}
	return result
}
