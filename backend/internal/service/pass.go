package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

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
	residentRepo    domain.ResidentRepository
	ruleRepo        domain.RuleRepository
	scanEventRepo   domain.ScanEventRepository
	personalPassKey []byte
	fallbackLogger  *zap.Logger
	opsTotal        *prometheus.CounterVec
	createdByType   *prometheus.CounterVec
	rejectionsTotal *prometheus.CounterVec
}

func NewPassService(
	passRepo domain.PassRepository,
	apartmentRepo domain.ApartmentRepository,
	residentRepo domain.ResidentRepository,
	ruleRepo domain.RuleRepository,
	scanEventRepo domain.ScanEventRepository,
	personalPassSecret string,
	logger *zap.Logger,
	m *metrics.Metrics,
) *PassService {
	if strings.TrimSpace(personalPassSecret) == "" {
		personalPassSecret = "yardpass-personal-pass-secret"
	}

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

	// Prime rejection label combinations so Grafana always sees the series (rate() works; no "No data").
	for _, reason := range []string{
		"invalid_car_plate",
		"max_duration_exceeded",
		"quiet_hours",
		"daily_limit_exceeded",
	} {
		rejectionsTotal.WithLabelValues(reason)
	}

	return &PassService{
		passRepo:        passRepo,
		apartmentRepo:   apartmentRepo,
		residentRepo:    residentRepo,
		ruleRepo:        ruleRepo,
		scanEventRepo:   scanEventRepo,
		personalPassKey: []byte(personalPassSecret),
		fallbackLogger:  logger,
		opsTotal:        opsTotal,
		createdByType:   createdByType,
		rejectionsTotal: rejectionsTotal,
	}
}

func (s *PassService) CreatePass(ctx context.Context, req domain.CreatePassRequest) (*domain.CreatePassResult, error) {
	var carPlate *string
	if req.CarPlate != nil && *req.CarPlate != "" {
		normalized := NormalizeCarPlate(*req.CarPlate)
		if normalized == "" {
			s.rejectionsTotal.WithLabelValues("invalid_car_plate").Inc()
			return nil, errors.New("Некорректный номер автомобиля.")
		}
		carPlate = &normalized
	}

	apartment, err := s.apartmentRepo.GetByID(ctx, req.ApartmentID)
	if err != nil {
		return nil, errors.New("Не удалось проверить квартиру. Попробуйте позже.")
	}
	if apartment == nil {
		return nil, errors.New("Квартира не найдена.")
	}

	rule, err := s.ruleRepo.GetByBuildingID(ctx, apartment.BuildingID)
	if err != nil {
		return nil, errors.New("Не удалось загрузить правила здания. Попробуйте позже.")
	}
	if rule == nil {
		rule = &domain.Rule{
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}
	}

	validFromUTC := req.ValidFrom.UTC()
	validToUTC := req.ValidTo.UTC()

	maxDuration := time.Duration(rule.MaxPassDurationHours) * time.Hour
	if validToUTC.Sub(validFromUTC) > maxDuration {
		s.rejectionsTotal.WithLabelValues("max_duration_exceeded").Inc()
		return nil, fmt.Errorf("Срок действия пропуска не может превышать %d ч.", rule.MaxPassDurationHours)
	}

	if req.ResidentID == nil {
		return nil, errors.New("Не указан житель (resident_id).")
	}

	resident, err := s.residentRepo.GetByID(ctx, *req.ResidentID)
	if err != nil {
		return nil, errors.New("Не удалось проверить жителя. Попробуйте позже.")
	}
	if resident == nil {
		return nil, errors.New("Житель не найден.")
	}
	if resident.ApartmentID != req.ApartmentID {
		return nil, errors.New("Житель не относится к указанной квартире.")
	}

	localLocation := locationForResidentRules(resident)

	var quietNotice *string
	if ruleQuietHoursConfigured(rule) {
		validFromLocal := validFromUTC.In(localLocation)
		validToLocal := validToUTC.In(localLocation)
		newToLocal, truncated, truncErr := truncateValidToForQuietHours(validFromLocal, validToLocal, *rule.QuietHoursStart, *rule.QuietHoursEnd)
		if truncErr != nil {
			s.rejectionsTotal.WithLabelValues("quiet_hours").Inc()
			return nil, truncErr
		}
		validToUTC = newToLocal.UTC()
		if truncated {
			msg := quietHoursShortenedNotice(newToLocal)
			quietNotice = &msg
		}
	}

	pass := &domain.Pass{
		ID:          uuid.New(),
		ApartmentID: req.ApartmentID,
		ResidentID:  req.ResidentID,
		CarPlate:    carPlate,
		GuestName:   req.GuestName,
		ValidFrom:   validFromUTC,
		ValidTo:     validToUTC,
		Status:      "active",
	}

	// Daily limit is enforced on the resident's local calendar day at pass creation time (not valid_from),
	// so limits apply correctly when valid_from differs from the creation instant (e.g. API clients).
	nowUTC := time.Now().UTC()
	dayAnchorLocal := nowUTC.In(localLocation)
	startOfDayLocal := time.Date(dayAnchorLocal.Year(), dayAnchorLocal.Month(), dayAnchorLocal.Day(), 0, 0, 0, 0, localLocation)
	endOfDayLocal := startOfDayLocal.Add(24 * time.Hour)

	created, err := s.passRepo.CreateWithDailyLimit(
		ctx,
		pass,
		startOfDayLocal.UTC(),
		endOfDayLocal.UTC(),
		int(rule.DailyPassLimitPerApartment),
	)
	if err != nil {
		s.opsTotal.WithLabelValues("create", "error").Inc()
		return nil, errors.New("Не удалось создать пропуск. Попробуйте позже.")
	}
	if !created {
		s.rejectionsTotal.WithLabelValues("daily_limit_exceeded").Inc()
		return nil, fmt.Errorf("Превышен дневной лимит пропусков для квартиры (не более %d в сутки).", rule.DailyPassLimitPerApartment)
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

	return &domain.CreatePassResult{Pass: pass, QuietHoursNotice: quietNotice}, nil
}

func (s *PassService) GenerateResidentPersonalPassToken(residentTelegramID int64) string {
	return fmt.Sprintf("resident:%d:%s", residentTelegramID, s.signResidentToken(residentTelegramID))
}

func (s *PassService) ValidatePass(ctx context.Context, passID uuid.UUID, guardUserID int64, buildingID *int64) (*domain.PassValidationResult, error) {
	pass, err := s.passRepo.GetByID(ctx, passID)
	if err != nil {
		return nil, errors.New("Не удалось проверить пропуск. Попробуйте позже.")
	}

	if pass == nil {
		result := &domain.PassValidationResult{
			Valid:  false,
			Reason: "PASS_NOT_FOUND",
		}
		return result, nil
	}

	if buildingID != nil {
		apartment, err := s.apartmentRepo.GetByID(ctx, pass.ApartmentID)
		if err != nil {
			return nil, errors.New("Не удалось проверить квартиру. Попробуйте позже.")
		}
		if apartment == nil {
			return &domain.PassValidationResult{Valid: false, Reason: "APARTMENT_NOT_FOUND"}, nil
		}
		if apartment.BuildingID != *buildingID {
			result := &domain.PassValidationResult{Valid: false, Reason: "BUILDING_MISMATCH"}
			s.opsTotal.WithLabelValues("validate", "invalid").Inc()
			s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
			return result, nil
		}
	}

	return s.validatePassInternal(ctx, pass, guardUserID)
}

func (s *PassService) ValidatePassByCarPlate(ctx context.Context, carPlate string, guardUserID int64, buildingID *int64) (*domain.PassValidationResult, error) {
	normalizedCarPlate := NormalizeCarPlate(carPlate)
	if normalizedCarPlate == "" {
		result := &domain.PassValidationResult{
			Valid:  false,
			Reason: "INVALID_CAR_PLATE",
		}
		return result, nil
	}

	pass, err := s.passRepo.GetActiveByCarPlate(ctx, normalizedCarPlate, buildingID)
	if err != nil {
		return nil, errors.New("Не удалось проверить пропуск по номеру. Попробуйте позже.")
	}

	if pass == nil {
		return s.validateResidentCarPlate(ctx, normalizedCarPlate, buildingID)
	}

	return s.validatePassInternal(ctx, pass, guardUserID)
}

// validateResidentCarPlate checks permanent resident vehicle registration when no guest pass matches.
func (s *PassService) validateResidentCarPlate(ctx context.Context, normalizedCarPlate string, buildingID *int64) (*domain.PassValidationResult, error) {
	residents, err := s.residentRepo.ListActiveWithCarPlate(ctx, buildingID)
	if err != nil {
		return nil, errors.New("Не удалось проверить данные жителей. Попробуйте позже.")
	}

	result := &domain.PassValidationResult{Valid: false}

	for i := range residents {
		res := &residents[i]
		if res.CarPlate == nil {
			continue
		}
		if NormalizeCarPlate(*res.CarPlate) != normalizedCarPlate {
			continue
		}

		apartment, err := s.apartmentRepo.GetByID(ctx, res.ApartmentID)
		if err != nil || apartment == nil {
			return &domain.PassValidationResult{Valid: false, Reason: "APARTMENT_NOT_FOUND"}, nil
		}
		if buildingID != nil && apartment.BuildingID != *buildingID {
			result.Reason = "BUILDING_MISMATCH"
			return result, nil
		}

		return &domain.PassValidationResult{
			Valid:     true,
			CarPlate:  normalizedCarPlate,
			Apartment: apartment.Number,
		}, nil
	}

	result.Reason = "PASS_NOT_FOUND"
	return result, nil
}

func (s *PassService) ValidateResidentPersonalPass(ctx context.Context, token string, guardUserID int64, buildingID *int64) (*domain.PassValidationResult, error) {
	// Trim whitespace that QR scanners sometimes append
	token = strings.TrimSpace(token)

	const prefix = "resident:"
	if !strings.HasPrefix(token, prefix) {
		return &domain.PassValidationResult{Valid: false, Reason: "INVALID_PERSONAL_PASS"}, nil
	}

	// Format: resident:<telegram_id>:<hmac>
	// Use SplitN to limit to 3 parts so any unexpected colons don't break parsing
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return &domain.PassValidationResult{Valid: false, Reason: "INVALID_PERSONAL_PASS"}, nil
	}

	telegramID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return &domain.PassValidationResult{Valid: false, Reason: "INVALID_PERSONAL_PASS"}, nil
	}

	expectedHMAC := s.signResidentToken(telegramID)
	if !hmac.Equal([]byte(strings.TrimSpace(parts[2])), []byte(expectedHMAC)) {
		return &domain.PassValidationResult{Valid: false, Reason: "INVALID_PERSONAL_PASS"}, nil
	}

	resident, err := s.residentRepo.GetByTelegramID(ctx, telegramID)
	if err != nil || resident == nil || resident.Status != "active" {
		return &domain.PassValidationResult{Valid: false, Reason: "RESIDENT_NOT_FOUND"}, nil
	}

	apartment, err := s.apartmentRepo.GetByID(ctx, resident.ApartmentID)
	if err != nil || apartment == nil {
		return &domain.PassValidationResult{Valid: false, Reason: "APARTMENT_NOT_FOUND"}, nil
	}

	if buildingID != nil && apartment.BuildingID != *buildingID {
		return &domain.PassValidationResult{Valid: false, Reason: "BUILDING_MISMATCH"}, nil
	}

	carPlate := ""
	if resident.CarPlate != nil && *resident.CarPlate != "" {
		carPlate = NormalizeCarPlate(*resident.CarPlate)
	}

	return &domain.PassValidationResult{
		Valid:     true,
		CarPlate:  carPlate,
		Apartment: apartment.Number,
	}, nil
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

	if pass.Status == "used" {
		result.Reason = "PASS_ALREADY_USED"
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
			if ruleQuietHoursConfigured(rule) {
				localLocation := s.quietHoursLocation(ctx, pass)
				if s.isQuietHours(now.In(localLocation), *rule.QuietHoursStart, *rule.QuietHoursEnd) {
					result.Reason = "QUIET_HOURS"
					s.opsTotal.WithLabelValues("validate", "invalid").Inc()
					s.logScanEvent(ctx, pass.ID, guardUserID, "invalid", result.Reason)
					return result, nil
				}
			}
		}
	}

	// First successful validation: mark pass as used
	pass.Status = "used"
	_ = s.passRepo.Update(ctx, pass)

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
		return errors.New("Не удалось найти пропуск. Попробуйте позже.")
	}
	if pass == nil {
		return errors.New("Пропуск не найден.")
	}

	if pass.Status == "revoked" {
		return errors.New("Пропуск уже отозван.")
	}

	if err := s.passRepo.Revoke(ctx, passID); err != nil {
		s.opsTotal.WithLabelValues("revoke", "error").Inc()
		return errors.New("Не удалось отозвать пропуск. Попробуйте позже.")
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
	normalized := NormalizeCarPlate(carPlate)
	if normalized == "" {
		return []domain.Pass{}, nil
	}
	return s.passRepo.SearchByCarPlate(ctx, normalized, buildingID, 50)
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

func (s *PassService) signResidentToken(telegramID int64) string {
	mac := hmac.New(sha256.New, s.personalPassKey)
	mac.Write([]byte(strconv.FormatInt(telegramID, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ruleQuietHoursConfigured is true when both fields are set and non-empty (after trim).
// If only one side is set in the DB, quiet-hours logic is skipped entirely — avoid relying on that in production.
func ruleQuietHoursConfigured(rule *domain.Rule) bool {
	if rule == nil {
		return false
	}
	if rule.QuietHoursStart == nil || rule.QuietHoursEnd == nil {
		return false
	}
	if strings.TrimSpace(*rule.QuietHoursStart) == "" || strings.TrimSpace(*rule.QuietHoursEnd) == "" {
		return false
	}
	return true
}

// intervalsOverlapHalfOpen reports whether [a0, a1) and [b0, b1) overlap.
func intervalsOverlapHalfOpen(a0, a1, b0, b1 time.Time) bool {
	return a0.Before(b1) && b0.Before(a1)
}

// truncateValidToForQuietHours shortens validTo to the start of the first quiet-hours window that intersects the pass interval.
// If the interval begins inside quiet hours, creation is rejected.
func truncateValidToForQuietHours(validFrom, validTo time.Time, startStr, endStr string) (newTo time.Time, truncated bool, err error) {
	startClock, err := parseTime(startStr)
	if err != nil {
		return time.Time{}, false, errors.New("Некорректное время начала тихих часов в правилах здания.")
	}
	endClock, err := parseTime(endStr)
	if err != nil {
		return time.Time{}, false, errors.New("Некорректное время окончания тихих часов в правилах здания.")
	}
	if !validTo.After(validFrom) {
		return time.Time{}, false, errors.New("Время окончания пропуска должно быть позже времени начала.")
	}
	loc := validFrom.Location()
	fromDay := time.Date(validFrom.Year(), validFrom.Month(), validFrom.Day(), 0, 0, 0, 0, loc)
	toDay := time.Date(validTo.Year(), validTo.Month(), validTo.Day(), 0, 0, 0, 0, loc)

	newTo = validTo
	for d := fromDay; !d.After(toDay); d = d.Add(24 * time.Hour) {
		q0, q1 := quietHoursWindowOnDay(d, startClock, endClock)
		if !intervalsOverlapHalfOpen(validFrom, newTo, q0, q1) {
			continue
		}
		if !validFrom.Before(q0) && validFrom.Before(q1) {
			return time.Time{}, false, errors.New("Сейчас действуют тихие часы. Оформите пропуск после их окончания.")
		}
		if validFrom.Before(q0) && q0.Before(validTo) && q0.Before(newTo) {
			newTo = q0
		}
	}
	if !newTo.After(validFrom) {
		return time.Time{}, false, errors.New("Срок действия пропуска слишком короткий относительно тихих часов. Выберите другое время.")
	}
	return newTo, newTo.Before(validTo), nil
}

func quietHoursShortenedNotice(validToLocal time.Time) string {
	return fmt.Sprintf(
		"С %s действуют тихие часы (вход по пропуску в это время недоступен). Пропуск действителен до %s.",
		validToLocal.Format("15:04"),
		validToLocal.Format("15:04 02.01.2006"),
	)
}

// quietHoursWindowOnDay returns [q0, q1) for the quiet-hours rule on local calendar day d (end may be next calendar day).
func quietHoursWindowOnDay(d time.Time, startClock, endClock time.Time) (q0, q1 time.Time) {
	loc := d.Location()
	startMin := startClock.Hour()*60 + startClock.Minute()
	endMin := endClock.Hour()*60 + endClock.Minute()
	q0 = time.Date(d.Year(), d.Month(), d.Day(), startClock.Hour(), startClock.Minute(), 0, 0, loc)
	if endMin <= startMin {
		q1 = time.Date(d.Year(), d.Month(), d.Day(), endClock.Hour(), endClock.Minute(), 0, 0, loc).Add(24 * time.Hour)
	} else {
		q1 = time.Date(d.Year(), d.Month(), d.Day(), endClock.Hour(), endClock.Minute(), 0, 0, loc)
		if !q1.After(q0) {
			q1 = q1.Add(24 * time.Hour)
		}
	}
	return q0, q1
}

func (s *PassService) isQuietHours(nowLocal time.Time, startTime, endTime string) bool {
	startClock, err := parseTime(startTime)
	if err != nil {
		return false
	}
	endClock, err := parseTime(endTime)
	if err != nil {
		return false
	}
	d := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, nowLocal.Location())
	q0, q1 := quietHoursWindowOnDay(d, startClock, endClock)
	return !nowLocal.Before(q0) && nowLocal.Before(q1)
}

func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04", timeStr)
}

// locationForResidentRules returns the wall-clock zone for interpreting rule quiet hours
// (from pass validity) and the resident's calendar day for daily pass creation limits.
// Pass validity is always stored in UTC; rules are edited in the resident's local time
// (IANA name on resident, default Europe/Moscow).
func locationForResidentRules(resident *domain.Resident) *time.Location {
	if resident != nil && resident.Timezone != nil {
		name := strings.TrimSpace(*resident.Timezone)
		if name != "" {
			if loc, err := time.LoadLocation(name); err == nil {
				return loc
			}
		}
	}
	return europeMoscowOrMSK()
}

func europeMoscowOrMSK() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*3600)
	}
	return loc
}

func (s *PassService) quietHoursLocation(ctx context.Context, pass *domain.Pass) *time.Location {
	if pass.ResidentID != nil {
		resident, err := s.residentRepo.GetByID(ctx, *pass.ResidentID)
		if err == nil && resident != nil {
			return locationForResidentRules(resident)
		}
		if err != nil {
			lgr := logger.FromContext(ctx)
			if lgr == nil {
				lgr = s.fallbackLogger
			}
			lgr.Warn("failed to load resident for quiet-hours zone; using default",
				zap.Error(err),
				zap.String("pass_id", pass.ID.String()),
			)
		}
	}
	return locationForResidentRules(nil)
}
