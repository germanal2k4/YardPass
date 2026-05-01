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
	const listResidentsWithApartmentNumberQuery = `
		SELECT r.id, r.apartment_id, a.number, r.telegram_id, r.chat_id, r.name, r.phone, r.status, r.created_at, r.updated_at
		FROM residents r
		INNER JOIN apartments a ON a.id = r.apartment_id
		WHERE ($1::bigint IS NULL OR r.apartment_id = $1)
		  AND ($2::bigint IS NULL OR a.building_id = $2)
		  AND ($3::varchar IS NULL OR r.status = $3)
		ORDER BY r.created_at DESC
		LIMIT $4
		OFFSET $5
	`

	rows, err := r.pool.Query(
		ctx,
		listResidentsWithApartmentNumberQuery,
		filters.ApartmentID,
		filters.BuildingID,
		filters.Status,
		intToInt32Ptr(filters.Limit),
		intToInt32Ptr(filters.Offset),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Resident
	for rows.Next() {
		var resident domain.Resident
		if err := rows.Scan(
			&resident.ID,
			&resident.ApartmentID,
			&resident.ApartmentNumber,
			&resident.TelegramID,
			&resident.ChatID,
			&resident.Name,
			&resident.Phone,
			&resident.Status,
			&resident.CreatedAt,
			&resident.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, resident)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ResidentRepo) Delete(ctx context.Context, id int64) error {
	ctx = queryNameToContext(ctx, "ResidentRepo.Delete")
	return r.queries.DeleteResident(ctx, id)
}

func residentFromDB(row db.Resident) *domain.Resident {
	return &domain.Resident{
		ID:          row.ID,
		ApartmentID: row.ApartmentID,
		TelegramID:  row.TelegramID,
		ChatID:      row.ChatID,
		Name:        row.Name,
		Phone:       row.Phone,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

