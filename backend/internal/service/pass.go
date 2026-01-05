package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"yardpass/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PassService struct {
	passRepo      domain.PassRepository
	apartmentRepo domain.ApartmentRepository
	ruleRepo      domain.RuleRepository
	scanEventRepo domain.ScanEventRepository
	logger        *zap.Logger
}

func NewPassService(
	passRepo domain.PassRepository,
	apartmentRepo domain.ApartmentRepository,
	ruleRepo domain.RuleRepository,
	scanEventRepo domain.ScanEventRepository,
	logger *zap.Logger,
) *PassService {
	return &PassService{
		passRepo:      passRepo,
		apartmentRepo: apartmentRepo,
		ruleRepo:      ruleRepo,
		scanEventRepo: scanEventRepo,
		logger:        logger,
	}
}

func (s *PassService) CreatePass(ctx context.Context, req domain.CreatePassRequest) (*domain.Pass, error) {
	var carPlate *string
	if req.CarPlate != nil && *req.CarPlate != "" {
		normalized := normalizeCarPlate(*req.CarPlate)
		if normalized == "" {
			return nil, errors.New("invalid car plate number")
		}
		carPlate = &normalized
	}
	// carPlate can be nil for pedestrian guests

	apartment, err := s.apartmentRepo.GetByID(ctx, req.ApartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartment: %w", err)
	}
	if apartment == nil {
		return nil, errors.New("apartment not found")
	}

	rule, err := s.ruleRepo.GetByBuildingID(ctx, apartment.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	if rule == nil {
		rule = &domain.Rule{
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}
	}

	maxDuration := time.Duration(rule.MaxPassDurationHours) * time.Hour
	if req.ValidTo.Sub(req.ValidFrom) > maxDuration {
		return nil, fmt.Errorf("pass duration exceeds maximum of %d hours", rule.MaxPassDurationHours)
	}

	// Проверяем лимит для конкретного жителя, а не для всей квартиры
	if req.ResidentID == nil {
		return nil, errors.New("resident_id is required")
	}

	count, err := s.passRepo.CountActiveTodayByResidentID(ctx, *req.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily limit: %w", err)
	}
	if count >= rule.DailyPassLimitPerApartment {
		return nil, fmt.Errorf("daily pass limit exceeded: you have created %d passes today (limit: %d)", count, rule.DailyPassLimitPerApartment)
	}

	if rule.QuietHoursStart != nil && rule.QuietHoursEnd != nil {
		if err := s.validateQuietHours(req.ValidFrom, req.ValidTo, *rule.QuietHoursStart, *rule.QuietHoursEnd); err != nil {
			return nil, err
		}
	}

	pass := &domain.Pass{
		ID:          uuid.New(),
		ApartmentID: req.ApartmentID,
		ResidentID:  req.ResidentID,
		CarPlate:    carPlate,
		GuestName:   req.GuestName,
		ValidFrom:   req.ValidFrom,
		ValidTo:     req.ValidTo,
		Status:      "active",
	}

	if err := s.passRepo.Create(ctx, pass); err != nil {
		return nil, fmt.Errorf("failed to create pass: %w", err)
	}

	logFields := []zap.Field{
		zap.String("pass_id", pass.ID.String()),
		zap.Int64("apartment_id", pass.ApartmentID),
	}
	if carPlate != nil {
		logFields = append(logFields, zap.String("car_plate", *carPlate))
	} else {
		logFields = append(logFields, zap.String("type", "pedestrian"))
	}
	s.logger.Info("pass created", logFields...)

	return pass, nil
}

func (s *PassService) ValidatePass(ctx context.Context, passID uuid.UUID, guardUserID int64) (*domain.PassValidationResult, error) {
	pass, err := s.passRepo.GetByID(ctx, passID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pass: %w", err)
	}

	result := &domain.PassValidationResult{
		Valid: false,
	}

	if pass == nil {
		result.Reason = "PASS_NOT_FOUND"
		s.logScanEvent(ctx, passID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	if pass.Status == "revoked" {
		result.Reason = "PASS_REVOKED"
		s.logScanEvent(ctx, passID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	now := time.Now()
	if now.Before(pass.ValidFrom) {
		result.Reason = "PASS_NOT_YET_VALID"
		s.logScanEvent(ctx, passID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	if now.After(pass.ValidTo) {
		result.Reason = "PASS_EXPIRED"
		pass.Status = "expired"
		_ = s.passRepo.Update(ctx, pass)
		s.logScanEvent(ctx, passID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	apartment, err := s.apartmentRepo.GetByID(ctx, pass.ApartmentID)
	if err == nil && apartment != nil {
		rule, err := s.ruleRepo.GetByBuildingID(ctx, apartment.BuildingID)
		if err == nil && rule != nil {
			if rule.QuietHoursStart != nil && rule.QuietHoursEnd != nil {
				if s.isQuietHours(now, *rule.QuietHoursStart, *rule.QuietHoursEnd) {
					result.Reason = "QUIET_HOURS"
					s.logScanEvent(ctx, passID, guardUserID, "invalid", result.Reason)
					return result, nil
				}
			}
		}
	}

	result.Valid = true
	result.Pass = pass
	if pass.CarPlate != nil {
		result.CarPlate = *pass.CarPlate
	}
	result.ValidTo = &pass.ValidTo

	if apartment != nil {
		result.Apartment = apartment.Number
	}

	s.logScanEvent(ctx, passID, guardUserID, "valid", "")
	return result, nil
}

func (s *PassService) RevokePass(ctx context.Context, passID uuid.UUID, revokedBy int64) error {
	pass, err := s.passRepo.GetByID(ctx, passID)
	if err != nil {
		return fmt.Errorf("failed to get pass: %w", err)
	}
	if pass == nil {
		return errors.New("pass not found")
	}

	if pass.Status == "revoked" {
		return errors.New("pass already revoked")
	}

	if err := s.passRepo.Revoke(ctx, passID); err != nil {
		return fmt.Errorf("failed to revoke pass: %w", err)
	}

	s.logger.Info("pass revoked",
		zap.String("pass_id", passID.String()),
		zap.Int64("revoked_by", revokedBy),
	)

	return nil
}

func (s *PassService) GetActivePasses(ctx context.Context, apartmentID int64) ([]*domain.Pass, error) {
	return s.passRepo.GetActiveByApartmentID(ctx, apartmentID)
}

func (s *PassService) GetActivePassesByResident(ctx context.Context, residentID int64) ([]*domain.Pass, error) {
	return s.passRepo.GetActiveByResidentID(ctx, residentID)
}

func (s *PassService) GetActivePassesByBuilding(ctx context.Context, buildingID int64) ([]*domain.Pass, error) {
	return s.passRepo.GetActiveByBuildingID(ctx, buildingID)
}

func (s *PassService) SearchPassesByCarPlate(ctx context.Context, carPlate string, buildingID *int64) ([]*domain.Pass, error) {
	return s.passRepo.SearchByCarPlate(ctx, carPlate, buildingID, 50)
}

func (s *PassService) logScanEvent(ctx context.Context, passID uuid.UUID, guardUserID int64, result, reason string) {
	event := &domain.ScanEvent{
		PassID:      passID,
		GuardUserID: guardUserID,
		ScannedAt:   time.Now(),
		Result:      result,
		Reason:      &reason,
	}

	if err := s.scanEventRepo.Create(ctx, event); err != nil {
		s.logger.Error("failed to log scan event", zap.Error(err))
	}
}

func normalizeCarPlate(plate string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(plate), " ", ""))

	var result strings.Builder
	for _, r := range normalized {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (s *PassService) validateQuietHours(validFrom, validTo time.Time, startTime, endTime string) error {
	start, err := parseTime(startTime)
	if err != nil {
		return fmt.Errorf("invalid quiet hours start: %w", err)
	}
	end, err := parseTime(endTime)
	if err != nil {
		return fmt.Errorf("invalid quiet hours end: %w", err)
	}

	fromHour := validFrom.Hour()*60 + validFrom.Minute()
	toHour := validTo.Hour()*60 + validTo.Minute()
	startMin := start.Hour()*60 + start.Minute()
	endMin := end.Hour()*60 + end.Minute()

	if endMin < startMin {
		endMin += 24 * 60
		if toHour < startMin {
			toHour += 24 * 60
		}
	}

	if (fromHour < endMin && toHour > startMin) || (fromHour+24*60 < endMin && toHour+24*60 > startMin) {
		return errors.New("pass cannot overlap with quiet hours")
	}

	return nil
}

func (s *PassService) isQuietHours(now time.Time, startTime, endTime string) bool {
	start, err := parseTime(startTime)
	if err != nil {
		return false
	}
	end, err := parseTime(endTime)
	if err != nil {
		return false
	}

	nowMin := now.Hour()*60 + now.Minute()
	startMin := start.Hour()*60 + start.Minute()
	endMin := end.Hour()*60 + end.Minute()

	if endMin < startMin {
		return nowMin >= startMin || nowMin < endMin
	}

	return nowMin >= startMin && nowMin < endMin
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04", timeStr)
}
