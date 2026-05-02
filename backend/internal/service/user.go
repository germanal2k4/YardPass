package service

import (
	"context"
	"errors"
	"strings"

	"yardpass/internal/auth"
	"yardpass/internal/domain"
	"yardpass/internal/observability/logger"
	"yardpass/internal/observability/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type UserService struct {
	userRepo       domain.UserRepository
	buildingRepo   domain.BuildingRepository
	ruleRepo       domain.RuleRepository
	fallbackLogger *zap.Logger
	opsTotal       *prometheus.CounterVec
}

func NewUserService(
	userRepo domain.UserRepository,
	buildingRepo domain.BuildingRepository,
	ruleRepo domain.RuleRepository,
	logger *zap.Logger,
	m *metrics.Metrics,
) *UserService {
	opsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_user",
			Name:      "operations_total",
			Help:      "Total number of user operations",
		},
		[]string{"operation", "result"},
	)

	m.GetRegistry().MustRegister(opsTotal)

	return &UserService{
		userRepo:       userRepo,
		buildingRepo:   buildingRepo,
		ruleRepo:       ruleRepo,
		fallbackLogger: logger,
		opsTotal:       opsTotal,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, req domain.RegisterUserRequest, createdBy int64) (*domain.User, error) {
	creator, err := s.userRepo.GetByID(ctx, createdBy)
	if err != nil {
		return nil, errors.New("Не удалось проверить учётную запись. Попробуйте позже.")
	}
	if creator == nil {
		return nil, errors.New("Инициатор регистрации не найден.")
	}

	if creator.Role != "superuser" && creator.Role != "admin" {
		return nil, errors.New("Регистрировать пользователей могут только суперпользователь или администратор.")
	}

	if req.Role == "superuser" {
		return nil, errors.New("Нельзя создать учётную запись суперпользователя.")
	}

	switch creator.Role {
	case "superuser":
		// superuser may create admin/guard for specific building
	case "admin":
		if req.Role != "guard" {
			return nil, errors.New("Администратор может создавать только учётные записи охраны.")
		}
		if creator.BuildingID == nil {
			return nil, errors.New("Администратор не привязан к зданию.")
		}
		if req.BuildingID != nil && *req.BuildingID != *creator.BuildingID {
			return nil, errors.New("Администратор может создавать пользователей только для своего здания.")
		}
		// Admin always creates guards in own building.
		req.BuildingID = creator.BuildingID
	default:
		return nil, errors.New("Недостаточно прав для регистрации пользователей.")
	}

	if req.Role == "guard" || req.Role == "admin" {
		if req.BuildingID == nil {
			if req.BuildingName == nil || strings.TrimSpace(*req.BuildingName) == "" {
				return nil, errors.New("Для роли охранник или администратор укажите building_id или название здания (building_name).")
			}
			resolvedBuildingID, err := s.resolveOrCreateBuildingByName(ctx, *req.BuildingName)
			if err != nil {
				return nil, err
			}
			req.BuildingID = &resolvedBuildingID
		}

		building, err := s.buildingRepo.GetByID(ctx, *req.BuildingID)
		if err != nil {
			return nil, errors.New("Не удалось проверить здание. Попробуйте позже.")
		}
		if building == nil {
			return nil, errors.New("Здание не найдено.")
		}

	}

	if req.ApartmentNumber != nil && *req.ApartmentNumber <= 0 {
		return nil, errors.New("Номер квартиры должен быть больше нуля.")
	}

	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("Не удалось проверить логин. Попробуйте позже.")
	}
	if existing != nil {
		return nil, errors.New("Пользователь с таким логином уже существует.")
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("Не удалось обработать пароль. Попробуйте другой пароль.")
	}

	user := &domain.User{
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    passwordHash,
		Role:            req.Role,
		BuildingID:      req.BuildingID,
		ApartmentNumber: req.ApartmentNumber,
		Status:          "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.opsTotal.WithLabelValues("register", "error").Inc()
		return nil, errors.New("Не удалось создать пользователя. Попробуйте позже.")
	}

	s.opsTotal.WithLabelValues("register", "success").Inc()

	lgr := logger.FromContext(ctx)
	if lgr == nil {
		lgr = s.fallbackLogger
	}
	lgr.Info("User registered",
		zap.String("username", user.Username),
		zap.String("role", user.Role),
		zap.Int64("created_by", createdBy),
	)

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, filters domain.UserFilters) ([]domain.User, error) {
	return s.userRepo.List(ctx, filters)
}

func (s *UserService) resolveOrCreateBuildingByName(ctx context.Context, rawName string) (int64, error) {
	normalized := normalizeBuildingName(rawName)
	if normalized == "" {
		return 0, errors.New("Укажите название здания (building_name).")
	}

	buildings, err := s.buildingRepo.List(ctx)
	if err != nil {
		return 0, errors.New("Не удалось загрузить список зданий. Попробуйте позже.")
	}
	for _, b := range buildings {
		if strings.EqualFold(normalizeBuildingName(b.Name), normalized) {
			return b.ID, nil
		}
	}

	building := &domain.Building{
		Name:           normalized,
		ApartmentCount: 1,
	}
	if err := s.buildingRepo.Create(ctx, building); err != nil {
		return 0, errors.New("Не удалось создать здание. Попробуйте позже.")
	}
	if err := s.ensureDefaultRule(ctx, building.ID); err != nil {
		return 0, err
	}
	return building.ID, nil
}

func normalizeBuildingName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func (s *UserService) ensureDefaultRule(ctx context.Context, buildingID int64) error {
	existingRule, err := s.ruleRepo.GetByBuildingID(ctx, buildingID)
	if err != nil {
		return errors.New("Не удалось загрузить правила здания. Попробуйте позже.")
	}
	if existingRule != nil {
		return nil
	}

	rule := &domain.Rule{
		BuildingID:                 buildingID,
		DailyPassLimitPerApartment: 5,
		MaxPassDurationHours:       24,
	}
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return errors.New("Не удалось создать правила по умолчанию. Попробуйте позже.")
	}
	return nil
}
