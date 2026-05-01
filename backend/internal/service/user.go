package service

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("get creator: %w", err)
	}
	if creator == nil {
		return nil, errors.New("creator not found")
	}

	if creator.Role != "superuser" && creator.Role != "admin" {
		return nil, errors.New("only superuser or admin can register users")
	}

	if req.Role == "superuser" {
		return nil, errors.New("cannot create superuser")
	}

	switch creator.Role {
	case "superuser":
		// superuser may create admin/guard for specific building
	case "admin":
		if req.Role != "guard" {
			return nil, errors.New("admin can only create guard accounts")
		}
		if creator.BuildingID == nil {
			return nil, errors.New("admin is not linked to a building")
		}
		if req.BuildingID != nil && *req.BuildingID != *creator.BuildingID {
			return nil, errors.New("admin can only create users for their own building")
		}
		// Admin always creates guards in own building.
		req.BuildingID = creator.BuildingID
	default:
		return nil, errors.New("insufficient permissions to register users")
	}

	if req.Role == "guard" || req.Role == "admin" {
		if req.BuildingID == nil {
			if req.BuildingName == nil || strings.TrimSpace(*req.BuildingName) == "" {
				return nil, errors.New("building_id or building_name is required for guard/admin")
			}
			resolvedBuildingID, err := s.resolveOrCreateBuildingByName(ctx, *req.BuildingName)
			if err != nil {
				return nil, err
			}
			req.BuildingID = &resolvedBuildingID
		}

		building, err := s.buildingRepo.GetByID(ctx, *req.BuildingID)
		if err != nil {
			return nil, fmt.Errorf("get building: %w", err)
		}
		if building == nil {
			return nil, errors.New("building not found")
		}

	}

	if req.ApartmentNumber != nil && *req.ApartmentNumber <= 0 {
		return nil, errors.New("apartment_number must be greater than zero")
	}

	existing, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
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
		return nil, fmt.Errorf("create user: %w", err)
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
		return 0, errors.New("building_name is required")
	}

	buildings, err := s.buildingRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list buildings: %w", err)
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
		return 0, fmt.Errorf("create building: %w", err)
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
		return fmt.Errorf("get building rule: %w", err)
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
		return fmt.Errorf("create default rule: %w", err)
	}
	return nil
}
