package service

import (
	"context"
	"testing"
	"time"

	"yardpass/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockPassRepo struct {
	mock.Mock
}

func (m *MockPassRepo) GetActiveByCarPlate(ctx context.Context, normalizedCarPlate string, buildingID *int64) (*domain.Pass, error) {
	args := m.Called(ctx, normalizedCarPlate, buildingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetActiveByResidentID(ctx context.Context, residentID int64) ([]*domain.Pass, error) {
	args := m.Called(ctx, residentID)
	return args.Get(0).([]*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pass, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetByApartmentID(ctx context.Context, apartmentID int64, status string) ([]*domain.Pass, error) {
	args := m.Called(ctx, apartmentID, status)
	return args.Get(0).([]*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetActiveByApartmentID(ctx context.Context, apartmentID int64) ([]*domain.Pass, error) {
	args := m.Called(ctx, apartmentID)
	return args.Get(0).([]*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) CountActiveTodayByApartmentID(ctx context.Context, apartmentID int64) (int, error) {
	args := m.Called(ctx, apartmentID)
	return args.Int(0), args.Error(1)
}

func (m *MockPassRepo) CountActiveTodayByResidentID(ctx context.Context, residentID int64) (int, error) {
	args := m.Called(ctx, residentID)
	return args.Int(0), args.Error(1)
}

func (m *MockPassRepo) Create(ctx context.Context, pass *domain.Pass) error {
	args := m.Called(ctx, pass)
	return args.Error(0)
}

func (m *MockPassRepo) Update(ctx context.Context, pass *domain.Pass) error {
	args := m.Called(ctx, pass)
	return args.Error(0)
}

func (m *MockPassRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPassRepo) GetActiveByBuildingID(ctx context.Context, buildingID int64) ([]*domain.Pass, error) {
	args := m.Called(ctx, buildingID)
	return args.Get(0).([]*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) SearchByCarPlate(ctx context.Context, carPlate string, buildingID *int64, limit int) ([]*domain.Pass, error) {
	args := m.Called(ctx, carPlate, buildingID, limit)
	return args.Get(0).([]*domain.Pass), args.Error(1)
}

type MockApartmentRepo struct {
	mock.Mock
}

func (m *MockApartmentRepo) GetByID(ctx context.Context, id int64) (*domain.Apartment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Apartment), args.Error(1)
}

func (m *MockApartmentRepo) GetByBuildingID(ctx context.Context, buildingID int64) ([]*domain.Apartment, error) {
	args := m.Called(ctx, buildingID)
	return args.Get(0).([]*domain.Apartment), args.Error(1)
}

func (m *MockApartmentRepo) GetByResidentTelegramID(ctx context.Context, telegramID int64) (*domain.Apartment, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Apartment), args.Error(1)
}

type MockRuleRepo struct {
	mock.Mock
}

func (m *MockRuleRepo) GetByBuildingID(ctx context.Context, buildingID int64) (*domain.Rule, error) {
	args := m.Called(ctx, buildingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *MockRuleRepo) Create(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepo) Update(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

type MockScanEventRepo struct {
	mock.Mock
}

func (m *MockScanEventRepo) Create(ctx context.Context, event *domain.ScanEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockScanEventRepo) List(ctx context.Context, filters domain.ScanEventFilters) ([]*domain.ScanEvent, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*domain.ScanEvent), args.Error(1)
}

func (m *MockScanEventRepo) CountValidScansToday(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func TestPassService_CreatePass(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	passRepo := new(MockPassRepo)
	apartmentRepo := new(MockApartmentRepo)
	ruleRepo := new(MockRuleRepo)
	scanEventRepo := new(MockScanEventRepo)

	service := NewPassService(passRepo, apartmentRepo, ruleRepo, scanEventRepo, logger)

	t.Run("successful creation", func(t *testing.T) {
		apartmentID := int64(1)
		buildingID := int64(1)
		now := time.Now()
		validTo := now.Add(2 * time.Hour)

		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
			Number:     "101",
		}, nil)

		ruleRepo.On("GetByBuildingID", ctx, buildingID).Return(&domain.Rule{
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}, nil)

		residentID := int64(1)
		passRepo.On("CountActiveTodayByResidentID", ctx, residentID).Return(2, nil)
		passRepo.On("Create", ctx, mock.AnythingOfType("*domain.Pass")).Return(nil)

		carPlate := "A123BC"
		req := domain.CreatePassRequest{
			ApartmentID: apartmentID,
			ResidentID:  &residentID,
			CarPlate:    &carPlate,
			ValidFrom:   now,
			ValidTo:     validTo,
		}

		pass, err := service.CreatePass(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, pass)
		assert.NotNil(t, pass.CarPlate)
		assert.Equal(t, "A123BC", *pass.CarPlate)
		assert.Equal(t, "active", pass.Status)

		passRepo.AssertExpectations(t)
		apartmentRepo.AssertExpectations(t)
		ruleRepo.AssertExpectations(t)
	})

	t.Run("daily limit exceeded", func(t *testing.T) {
		// Создаем новые моки для этого теста
		passRepo2 := new(MockPassRepo)
		apartmentRepo2 := new(MockApartmentRepo)
		ruleRepo2 := new(MockRuleRepo)
		scanEventRepo2 := new(MockScanEventRepo)
		service2 := NewPassService(passRepo2, apartmentRepo2, ruleRepo2, scanEventRepo2, logger)

		apartmentID := int64(1)
		buildingID := int64(1)
		now := time.Now()
		validTo := now.Add(2 * time.Hour)

		apartmentRepo2.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
		}, nil)

		ruleRepo2.On("GetByBuildingID", ctx, buildingID).Return(&domain.Rule{
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}, nil)

		// Лимит 5, уже создано 5, значит лимит превышен (>=)
		residentID := int64(1)
		passRepo2.On("CountActiveTodayByResidentID", ctx, residentID).Return(5, nil)

		carPlate := "A123BC"
		req := domain.CreatePassRequest{
			ApartmentID: apartmentID,
			ResidentID:  &residentID,
			CarPlate:    &carPlate,
			ValidFrom:   now,
			ValidTo:     validTo,
		}

		pass, err := service2.CreatePass(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, pass)
		if err != nil {
			assert.Contains(t, err.Error(), "daily pass limit")
		}

		passRepo2.AssertExpectations(t)
		apartmentRepo2.AssertExpectations(t)
		ruleRepo2.AssertExpectations(t)
	})
}

func TestPassService_ValidatePass(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	passRepo := new(MockPassRepo)
	apartmentRepo := new(MockApartmentRepo)
	ruleRepo := new(MockRuleRepo)
	scanEventRepo := new(MockScanEventRepo)

	service := NewPassService(passRepo, apartmentRepo, ruleRepo, scanEventRepo, logger)

	t.Run("valid pass", func(t *testing.T) {
		passID := uuid.New()
		apartmentID := int64(1)
		buildingID := int64(1)
		now := time.Now()

		carPlate := "A123BC"
		pass := &domain.Pass{
			ID:          passID,
			ApartmentID: apartmentID,
			CarPlate:    &carPlate,
			Status:      "active",
			ValidFrom:   now.Add(-1 * time.Hour),
			ValidTo:     now.Add(1 * time.Hour),
		}

		passRepo.On("GetByID", ctx, passID).Return(pass, nil)
		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
			Number:     "101",
		}, nil)
		ruleRepo.On("GetByBuildingID", ctx, buildingID).Return(&domain.Rule{}, nil)
		scanEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		result, err := service.ValidatePass(ctx, passID, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
	})

	t.Run("expired pass", func(t *testing.T) {
		passID := uuid.New()
		now := time.Now()

		pass := &domain.Pass{
			ID:          passID,
			ApartmentID: 1,
			Status:      "active",
			ValidFrom:   now.Add(-2 * time.Hour),
			ValidTo:     now.Add(-1 * time.Hour),
		}

		passRepo.On("GetByID", ctx, passID).Return(pass, nil)
		passRepo.On("Update", ctx, mock.AnythingOfType("*domain.Pass")).Return(nil)
		scanEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		result, err := service.ValidatePass(ctx, passID, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Equal(t, "PASS_EXPIRED", result.Reason)
	})
}
