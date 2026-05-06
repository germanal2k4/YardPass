package repo

import (
	"context"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/repo/db"
)

type ScanEventRepo struct {
	*PostgresRepo
}

func NewScanEventRepo(repo *PostgresRepo) *ScanEventRepo {
	return &ScanEventRepo{repo}
}

func (r *ScanEventRepo) Create(ctx context.Context, event *domain.ScanEvent) error {
	ctx = queryNameToContext(ctx, "ScanEventRepo.Create")

	id, err := r.queries.CreateScanEvent(ctx, db.CreateScanEventParams{
		PassID:      event.PassID,
		GuardUserID: event.GuardUserID,
		ScannedAt:   event.ScannedAt,
		Result:      event.Result,
		Reason:      event.Reason,
		Meta:        event.Meta,
	})
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func (r *ScanEventRepo) List(ctx context.Context, filters domain.ScanEventFilters) ([]domain.ScanEvent, error) {
	ctx = queryNameToContext(ctx, "ScanEventRepo.List")
	rows, err := r.queries.ListScanEvents(ctx, db.ListScanEventsParams{
		FilterPassID:      uuidToPgtype(filters.PassID),
		FilterGuardUserID: filters.GuardUserID,
		FilterResult:      filters.Result,
		FilterFrom:        timeToPgtypeTimestamp(filters.From),
		FilterTo:          timeToPgtypeTimestamp(filters.To),
		MaxResults:        intToInt32Ptr(filters.Limit),
		ResultsOffset:     intToInt32Ptr(filters.Offset),
	})
	if err != nil {
		return nil, err
	}
	return scanEventsFromDB(rows), nil
}

func (r *ScanEventRepo) CountValidScansToday(ctx context.Context) (int, error) {
	ctx = queryNameToContext(ctx, "ScanEventRepo.CountValidScansToday")
	today := time.Now().Truncate(24 * time.Hour)
	count, err := r.queries.CountValidScansToday(ctx, today)
	return int(count), err
}

func (r *ScanEventRepo) GetStatistics(ctx context.Context, from, to *time.Time, buildingID *int64) (*domain.Statistics, error) {
	ctx = queryNameToContext(ctx, "ScanEventRepo.GetStatistics")
	row, err := r.queries.GetScanEventStatistics(ctx, db.GetScanEventStatisticsParams{
		FilterFrom:       timeToPgtypeTimestamp(from),
		FilterTo:         timeToPgtypeTimestamp(to),
		FilterBuildingID: buildingID,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Statistics{
		TotalScans:   int(row.TotalScans),
		ValidScans:   int(row.ValidScans),
		InvalidScans: int(row.InvalidScans),
		UniquePasses: int(row.UniquePasses),
		UniqueGuards: int(row.UniqueGuards),
	}, nil
}

func (r *ScanEventRepo) GetEventsWithDetails(ctx context.Context, filters domain.ScanEventFilters, buildingID *int64) ([]domain.ScanEventWithDetails, error) {
	ctx = queryNameToContext(ctx, "ScanEventRepo.GetEventsWithDetails")
	rows, err := r.queries.GetScanEventsWithDetails(ctx, db.GetScanEventsWithDetailsParams{
		FilterBuildingID:      buildingID,
		FilterPassID:          uuidToPgtype(filters.PassID),
		FilterGuardUserID:     filters.GuardUserID,
		FilterResult:          filters.Result,
		FilterApartmentNumber: filters.ApartmentNumber,
		FilterCarPlate:        filters.CarPlate,
		FilterFrom:            timeToPgtypeTimestamp(filters.From),
		FilterTo:              timeToPgtypeTimestamp(filters.To),
		MaxResults:            intToInt32Ptr(filters.Limit),
		ResultsOffset:         intToInt32Ptr(filters.Offset),
	})
	if err != nil {
		return nil, err
	}
	return scanEventsWithDetailsFromDB(rows), nil
}

func scanEventsFromDB(rows []db.ScanEvent) []domain.ScanEvent {
	result := make([]domain.ScanEvent, len(rows))
	for i, e := range rows {
		result[i] = domain.ScanEvent(e)
	}
	return result
}

func scanEventsWithDetailsFromDB(rows []db.GetScanEventsWithDetailsRow) []domain.ScanEventWithDetails {
	result := make([]domain.ScanEventWithDetails, len(rows))
	for i, e := range rows {
		result[i] = domain.ScanEventWithDetails{
			ID:              e.ID,
			PassID:          e.PassID,
			GuardUserID:     e.GuardUserID,
			GuardUsername:   e.GuardUsername,
			ScannedAt:       e.ScannedAt,
			Result:          e.Result,
			Reason:          e.Reason,
			Meta:            e.Meta,
			CarPlate:        e.CarPlate,
			ApartmentNumber: e.ApartmentNumber,
			BuildingID:      e.BuildingID,
		}
	}
	return result
}
