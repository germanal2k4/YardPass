package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/observability/logger"
	"yardpass/internal/observability/metrics"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type PassService struct {
	passRepo        domain.PassRepository
	apartmentRepo   domain.ApartmentRepository
	ruleRepo        domain.RuleRepository
	scanEventRepo   domain.ScanEventRepository
	fallbackLogger  *zap.Logger
	opsTotal        *prometheus.CounterVec
	createdByType   *prometheus.CounterVec
	rejectionsTotal *prometheus.CounterVec
}

func NewPassService(
	passRepo domain.PassRepository,
	apartmentRepo domain.ApartmentRepository,
	ruleRepo domain.RuleRepository,
	scanEventRepo domain.ScanEventRepository,
	logger *zap.Logger,
	m *metrics.Metrics,
) *PassService {
	opsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_pass_service",
			Name:      "operations_total",
			Help:      "Total number of pass operations",
		},
		[]string{"operation", "result"},
	)

	createdByType := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_pass_service",
			Name:      "created_by_type_total",
			Help:      "Total number of passes created by guest type",
		},
		[]string{"type"},
	)

	rejectionsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_pass_service",
			Name:      "rejections_total",
			Help:      "Total number of pass creation rejections by reason",
		},
		[]string{"reason"},
	)

	m.GetRegistry().MustRegister(opsTotal, createdByType, rejectionsTotal)

	return &PassService{
		passRepo:        passRepo,
		apartmentRepo:   apartmentRepo,
		ruleRepo:        ruleRepo,
		scanEventRepo:   scanEventRepo,
		fallbackLogger:  logger,
		opsTotal:        opsTotal,
		createdByType:   createdByType,
		rejectionsTotal: rejectionsTotal,
	}
}

func (s *PassService) CreatePass(ctx context.Context, req domain.CreatePassRequest) (*domain.Pass, error) {
	var carPlate *string
	if req.CarPlate != nil && *req.CarPlate != "" {
		normalized := normalizeCarPlate(*req.CarPlate)
		if normalized == "" {
			s.rejectionsTotal.WithLabelValues("invalid_car_plate").Inc()
			return nil, errors.New("invalid car plate number")
		}
		carPlate = &normalized
	}

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
		s.rejectionsTotal.WithLabelValues("max_duration_exceeded").Inc()
		return nil, fmt.Errorf("pass duration exceeds maximum of %d hours", rule.MaxPassDurationHours)
	}

	if req.ResidentID == nil {
		return nil, errors.New("resident_id is required")
	}

	count, err := s.passRepo.CountActiveTodayByResidentID(ctx, *req.ResidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily limit: %w", err)
	}
	if count >= int(rule.DailyPassLimitPerApartment) {
		s.rejectionsTotal.WithLabelValues("daily_limit_exceeded").Inc()
		return nil, fmt.Errorf("daily pass limit exceeded: you have created %d passes today (limit: %d)", count, rule.DailyPassLimitPerApartment)
	}

	if rule.QuietHoursStart != nil && rule.QuietHoursEnd != nil {
		if err := s.validateQuietHours(req.ValidFrom, req.ValidTo, *rule.QuietHoursStart, *rule.QuietHoursEnd); err != nil {
			s.rejectionsTotal.WithLabelValues("quiet_hours").Inc()
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
		s.opsTotal.WithLabelValues("create", "error").Inc()
		return nil, fmt.Errorf("failed to create pass: %w", err)
	}

	s.opsTotal.WithLabelValues("create", "success").Inc()
	if carPlate != nil {
		s.createdByType.WithLabelValues("car").Inc()
	} else {
		s.createdByType.WithLabelValues("pedestrian").Inc()
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

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = s.fallbackLogger
	}
	lgr.Info("Pass created", logFields...)

	return pass, nil
}

func (s *PassService) ValidatePass(ctx context.Context, passID uuid.UUID, guardUserID int64) (*domain.PassValidationResult, error) {
	pass, err := s.passRepo.GetByID(ctx, passID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pass: %w", err)
	}

	if pass == nil {
		result := &domain.PassValidationResult{
			Valid:  false,
			Reason: "PASS_NOT_FOUND",
		}
		return result, nil
	}

	return s.validatePassInternal(ctx, pass, guardUserID)
}

func (s *PassService) ValidatePassByCarPlate(ctx context.Context, carPlate string, guardUserID int64, buildingID *int64) (*domain.PassValidationResult, error) {
	normalizedCarPlate := normalizeCarPlate(carPlate)
	if normalizedCarPlate == "" {
		result := &domain.PassValidationResult{
			Valid:  false,
			Reason: "INVALID_CAR_PLATE",
		}
		return result, nil
	}

	pass, err := s.passRepo.GetActiveByCarPlate(ctx, normalizedCarPlate, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pass by car plate: %w", err)
	}

	result := &domain.PassValidationResult{
		Valid: false,
	}

	if pass == nil {
		result.Reason = "PASS_NOT_FOUND"
		return result, nil
	}

	return s.validatePassInternal(ctx, pass, guardUserID)
}

func (s *PassService) validatePassInternal(ctx context.Context, pass *domain.Pass, guardUserID int64) (*domain.PassValidationResult, error) {
	result := &domain.PassValidationResult{
		Valid: false,
	}

	if pass.Status == "revoked" {
		result.Reason = "PASS_REVOKED"
		s.opsTotal.WithLabelValues("validate", "invalid").Inc()
		s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	now := time.Now().UTC()
	validFrom := pass.ValidFrom.UTC()
	validTo := pass.ValidTo.UTC()

	if now.Before(validFrom) {
		result.Reason = "PASS_NOT_YET_VALID"
		s.opsTotal.WithLabelValues("validate", "invalid").Inc()
		s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	if now.After(validTo) {
		result.Reason = "PASS_EXPIRED"
		pass.Status = "expired"
		_ = s.passRepo.Update(ctx, pass)
		s.opsTotal.WithLabelValues("validate", "invalid").Inc()
		s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
		return result, nil
	}

	apartment, err := s.apartmentRepo.GetByID(ctx, pass.ApartmentID)
	if err == nil && apartment != nil {
		rule, err := s.ruleRepo.GetByBuildingID(ctx, apartment.BuildingID)
		if err == nil && rule != nil {
			if rule.QuietHoursStart != nil && rule.QuietHoursEnd != nil {
				if s.isQuietHours(now, *rule.QuietHoursStart, *rule.QuietHoursEnd) {
					result.Reason = "QUIET_HOURS"
					s.opsTotal.WithLabelValues("validate", "invalid").Inc()
					s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
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

	s.opsTotal.WithLabelValues("validate", "valid").Inc()
	s.logScanEvent(ctx, pass.ID, guardUserID, "valid", "")
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
		s.opsTotal.WithLabelValues("revoke", "error").Inc()
		return fmt.Errorf("failed to revoke pass: %w", err)
	}

	s.opsTotal.WithLabelValues("revoke", "success").Inc()

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = s.fallbackLogger
	}

	lgr.Info("Pass revoked",
		zap.String("pass_id", passID.String()),
		zap.Int64("revoked_by", revokedBy),
	)

	return nil
}

func (s *PassService) GetActivePasses(ctx context.Context, apartmentID int64) ([]domain.Pass, error) {
	return s.passRepo.GetActiveByApartmentID(ctx, apartmentID)
}

func (s *PassService) GetActivePassesByResident(ctx context.Context, residentID int64) ([]domain.Pass, error) {
	return s.passRepo.GetActiveByResidentID(ctx, residentID)
}

func (s *PassService) GetActivePassesByBuilding(ctx context.Context, buildingID int64) ([]domain.Pass, error) {
	return s.passRepo.GetActiveByBuildingID(ctx, buildingID)
}

func (s *PassService) SearchPassesByCarPlate(ctx context.Context, carPlate string, buildingID *int64) ([]domain.Pass, error) {
	return s.passRepo.SearchByCarPlate(ctx, carPlate, buildingID, 50)
}

func (s *PassService) logScanEvent(ctx context.Context, passID uuid.UUID, guardUserID int64, result, reason string) {

	event := &domain.ScanEvent{
		PassID:      passID,
		GuardUserID: guardUserID,
		ScannedAt:   time.Now().UTC(),
		Result:      result,
		Reason:      &reason,
	}

	if err := s.scanEventRepo.Create(ctx, event); err != nil {
		lgr := logger.FromContext(ctx)
		if lgr == nil {
			lgr = s.fallbackLogger
		}

		lgr.Error("Failed to log scan event",
			zap.Error(err),
			zap.String("pass_id", passID.String()),
			zap.String("result", result),
			zap.String("reason", reason),
		)
	}
}

var russianToEnglish = map[rune]rune{
	'А': 'A', 'В': 'B', 'С': 'C', 'Е': 'E', 'К': 'K',
	'М': 'M', 'Н': 'H', 'О': 'O', 'Р': 'P', 'Т': 'T',
	'У': 'Y', 'Х': 'X',
	'а': 'A', 'в': 'B', 'с': 'C', 'е': 'E', 'к': 'K',
	'м': 'M', 'н': 'H', 'о': 'O', 'р': 'P', 'т': 'T',
	'у': 'Y', 'х': 'X',
}

func normalizeCarPlate(plate string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(plate), " ", ""))

	var result strings.Builder
	for _, r := range normalized {
		if eng, ok := russianToEnglish[r]; ok {
			result.WriteRune(eng)
		} else if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
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
