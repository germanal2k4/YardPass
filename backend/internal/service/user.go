package service

import (
	"context"
	"errors"
	"fmt"

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
	fallbackLogger *zap.Logger
	opsTotal       *prometheus.CounterVec
}

func NewUserService(userRepo domain.UserRepository, buildingRepo domain.BuildingRepository, logger *zap.Logger, m *metrics.Metrics) *UserService {
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

	if req.Role == "guard" || req.Role == "admin" {
		if req.BuildingID == nil {
			return nil, errors.New("building_id is required for guard/admin")
		}

		building, err := s.buildingRepo.GetByID(ctx, *req.BuildingID)
		if err != nil {
			return nil, fmt.Errorf("get building: %w", err)
		}
		if building == nil {
			return nil, errors.New("building not found")
		}

		if creator.Role == "admin" {
			if creator.BuildingID == nil || *creator.BuildingID != *req.BuildingID {
				return nil, errors.New("admin can only create users for their own building")
			}
		}
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
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		BuildingID:   req.BuildingID,
		Status:       "active",
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
