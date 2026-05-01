package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"

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

func (m *MockPassRepo) GetActiveByResidentID(ctx context.Context, residentID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, residentID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Pass, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetByApartmentID(ctx context.Context, apartmentID int64, status string) ([]domain.Pass, error) {
	args := m.Called(ctx, apartmentID, status)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *MockPassRepo) GetActiveByApartmentID(ctx context.Context, apartmentID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, apartmentID)
	return args.Get(0).([]domain.Pass), args.Error(1)
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

func (m *MockPassRepo) CreateWithDailyLimit(ctx context.Context, pass *domain.Pass, dayStartUTC, dayEndUTC time.Time, dailyLimit int) (bool, error) {
	args := m.Called(ctx, pass, dayStartUTC, dayEndUTC, dailyLimit)
	return args.Bool(0), args.Error(1)
}

func (m *MockPassRepo) Update(ctx context.Context, pass *domain.Pass) error {
	args := m.Called(ctx, pass)
	return args.Error(0)
}

func (m *MockPassRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPassRepo) GetActiveByBuildingID(ctx context.Context, buildingID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, buildingID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *MockPassRepo) SearchByCarPlate(ctx context.Context, carPlate string, buildingID *int64, limit int) ([]domain.Pass, error) {
	args := m.Called(ctx, carPlate, buildingID, limit)
	return args.Get(0).([]domain.Pass), args.Error(1)
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

func (m *MockApartmentRepo) GetByBuildingID(ctx context.Context, buildingID int64) ([]domain.Apartment, error) {
	args := m.Called(ctx, buildingID)
	return args.Get(0).([]domain.Apartment), args.Error(1)
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

type MockResidentRepo struct {
	mock.Mock
}

func (m *MockResidentRepo) GetByID(ctx context.Context, id int64) (*domain.Resident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *MockResidentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Resident, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *MockResidentRepo) SetCarPlate(ctx context.Context, id int64, carPlate *string) error {
	args := m.Called(ctx, id, carPlate)
	return args.Error(0)
}

func (m *MockResidentRepo) Create(ctx context.Context, resident *domain.Resident) error {
	args := m.Called(ctx, resident)
	return args.Error(0)
}

func (m *MockResidentRepo) Update(ctx context.Context, resident *domain.Resident) error {
	args := m.Called(ctx, resident)
	return args.Error(0)
}

func (m *MockResidentRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockResidentRepo) BulkCreate(ctx context.Context, residents []domain.Resident) error {
	args := m.Called(ctx, residents)
	return args.Error(0)
}

func (m *MockResidentRepo) List(ctx context.Context, filters domain.ResidentFilters) ([]domain.Resident, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.Resident), args.Error(1)
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

func (m *MockScanEventRepo) List(ctx context.Context, filters domain.ScanEventFilters) ([]domain.ScanEvent, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.ScanEvent), args.Error(1)
}

func (m *MockScanEventRepo) CountValidScansToday(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockScanEventRepo) GetEventsWithDetails(ctx context.Context, filters domain.ScanEventFilters, buildingID *int64) ([]domain.ScanEventWithDetails, error) {
	args := m.Called(ctx, filters, buildingID)
	return args.Get(0).([]domain.ScanEventWithDetails), args.Error(1)
}

func (m *MockScanEventRepo) GetStatistics(ctx context.Context, from *time.Time, to *time.Time, buildingID *int64) (*domain.Statistics, error) {
	args := m.Called(ctx, from, to, buildingID)
	return args.Get(0).(*domain.Statistics), args.Error(1)
}

func TestPassService_CreatePass(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	passRepo := new(MockPassRepo)
	apartmentRepo := new(MockApartmentRepo)
	ruleRepo := new(MockRuleRepo)
	residentRepo := new(MockResidentRepo)
	scanEventRepo := new(MockScanEventRepo)

	noopMetrics := &metrics.Metrics{}
	service := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, "test-secret", logger, noopMetrics)

	t.Run("successful creation", func(t *testing.T) {
		apartmentID := int64(1)
		buildingID := int64(1)
		now := time.Now()
		validTo := now.Add(2 * time.Hour)

		apartmentRepo.On("GetByID", mock.Anything, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
			Number:     "101",
		}, nil)

		ruleRepo.On("GetByBuildingID", ctx, buildingID).Return(&domain.Rule{
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}, nil)

		residentID := int64(1)
		passRepo.On("CreateWithDailyLimit", ctx, mock.AnythingOfType("*domain.Pass"), mock.Anything, mock.Anything, 5).Return(true, nil)

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
		passRepo2 := new(MockPassRepo)
		apartmentRepo2 := new(MockApartmentRepo)
		ruleRepo2 := new(MockRuleRepo)
		residentRepo2 := new(MockResidentRepo)
		scanEventRepo2 := new(MockScanEventRepo)
		service2 := NewPassService(passRepo2, apartmentRepo2, residentRepo2, ruleRepo2, scanEventRepo2, "test-secret", logger, noopMetrics)

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

		passRepo2.On("CreateWithDailyLimit", ctx, mock.AnythingOfType("*domain.Pass"), mock.Anything, mock.Anything, 5).Return(false, nil)

		residentID := int64(1)
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
	residentRepo := new(MockResidentRepo)
	scanEventRepo := new(MockScanEventRepo)
	noopMetrics := &metrics.Metrics{}

	service := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, "test-secret", logger, noopMetrics)

	t.Run("valid pass", func(t *testing.T) {
		passID := uuid.New()
		apartmentID := int64(1)
		buildingID := int64(1)
		guardBuildingID := int64(1)
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
		passRepo.On("Update", ctx, mock.AnythingOfType("*domain.Pass")).Return(nil)
		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
			Number:     "101",
		}, nil).Times(2)
		ruleRepo.On("GetByBuildingID", mock.Anything, buildingID).Return(&domain.Rule{}, nil)
		scanEventRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		result, err := service.ValidatePass(ctx, passID, 1, &guardBuildingID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
	})

	t.Run("expired pass", func(t *testing.T) {
		passID := uuid.New()
		now := time.Now()
		guardBuildingID := int64(1)

		pass := &domain.Pass{
			ID:          passID,
			ApartmentID: 1,
			Status:      "active",
			ValidFrom:   now.Add(-2 * time.Hour),
			ValidTo:     now.Add(-1 * time.Hour),
		}

		passRepo.On("GetByID", ctx, passID).Return(pass, nil)
		apartmentRepo.On("GetByID", ctx, int64(1)).Return(&domain.Apartment{
			ID:         1,
			BuildingID: guardBuildingID,
		}, nil)
		passRepo.On("Update", ctx, mock.AnythingOfType("*domain.Pass")).Return(nil)
		scanEventRepo.On("Create", ctx, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		result, err := service.ValidatePass(ctx, passID, 1, &guardBuildingID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Equal(t, "PASS_EXPIRED", result.Reason)
	})

	t.Run("cannot validate pass twice", func(t *testing.T) {
		passID := uuid.New()
		apartmentID := int64(1)
		buildingID := int64(1)
		guardBuildingID := int64(1)
		now := time.Now()

		pass := &domain.Pass{
			ID:          passID,
			ApartmentID: apartmentID,
			Status:      "active",
			ValidFrom:   now.Add(-1 * time.Hour),
			ValidTo:     now.Add(1 * time.Hour),
		}

		passRepo.On("GetByID", mock.Anything, passID).Return(pass, nil)
		passRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Pass")).Return(nil)
		apartmentRepo.On("GetByID", mock.Anything, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: buildingID,
			Number:     "101",
		}, nil).Times(3)
		ruleRepo.On("GetByBuildingID", mock.Anything, buildingID).Return(&domain.Rule{}, nil)
		scanEventRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		// First validation should be successful
		firstResult, err := service.ValidatePass(ctx, passID, 1, &guardBuildingID)
		assert.NoError(t, err)
		assert.NotNil(t, firstResult)
		assert.True(t, firstResult.Valid)

		// Second validation should be rejected as already used
		secondResult, err := service.ValidatePass(ctx, passID, 1, &guardBuildingID)
		assert.NoError(t, err)
		assert.NotNil(t, secondResult)
		assert.False(t, secondResult.Valid)
		assert.Equal(t, "PASS_ALREADY_USED", secondResult.Reason)
	})

	t.Run("building mismatch rejects without consuming pass", func(t *testing.T) {
		passID := uuid.New()
		apartmentID := int64(1)
		passBuildingID := int64(1)
		guardBuildingID := int64(2)
		now := time.Now()

		pass := &domain.Pass{
			ID:          passID,
			ApartmentID: apartmentID,
			Status:      "active",
			ValidFrom:   now.Add(-1 * time.Hour),
			ValidTo:     now.Add(1 * time.Hour),
		}

		passRepo.On("GetByID", ctx, passID).Return(pass, nil)
		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{
			ID:         apartmentID,
			BuildingID: passBuildingID,
			Number:     "101",
		}, nil)
		scanEventRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.ScanEvent")).Return(nil)

		result, err := service.ValidatePass(ctx, passID, 1, &guardBuildingID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Equal(t, "BUILDING_MISMATCH", result.Reason)
		passRepo.AssertNotCalled(t, "Update")
	})
}

func TestPassService_ValidateResidentPersonalPass(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()
	const secret = "test-personal-pass-secret"

	noopMetrics := &metrics.Metrics{}

	newService := func() *PassService {
		return NewPassService(
			new(MockPassRepo),
			new(MockApartmentRepo),
			new(MockResidentRepo),
			new(MockRuleRepo),
			new(MockScanEventRepo),
			secret,
			logger,
			noopMetrics,
		)
	}

	t.Run("valid personal pass returns apartment and car plate", func(t *testing.T) {
		passRepo := new(MockPassRepo)
		apartmentRepo := new(MockApartmentRepo)
		residentRepo := new(MockResidentRepo)
		ruleRepo := new(MockRuleRepo)
		scanEventRepo := new(MockScanEventRepo)
		svc := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, secret, logger, noopMetrics)

		telegramID := int64(123456789)
		carPlate := "A123BC"
		token := svc.GenerateResidentPersonalPassToken(telegramID)

		buildingID := int64(1)
		residentRepo.On("GetByTelegramID", ctx, telegramID).Return(&domain.Resident{
			ID:          1,
			ApartmentID: 10,
			TelegramID:  telegramID,
			CarPlate:    &carPlate,
			Status:      "active",
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(10)).Return(&domain.Apartment{
			ID:         10,
			BuildingID: buildingID,
			Number:     "42",
		}, nil)

		result, err := svc.ValidateResidentPersonalPass(ctx, token, 0, &buildingID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Equal(t, "42", result.Apartment)
		assert.Equal(t, carPlate, result.CarPlate)
	})

	t.Run("valid personal pass without car plate", func(t *testing.T) {
		passRepo := new(MockPassRepo)
		apartmentRepo := new(MockApartmentRepo)
		residentRepo := new(MockResidentRepo)
		ruleRepo := new(MockRuleRepo)
		scanEventRepo := new(MockScanEventRepo)
		svc := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, secret, logger, noopMetrics)

		telegramID := int64(987654321)
		token := svc.GenerateResidentPersonalPassToken(telegramID)

		buildingID := int64(1)
		residentRepo.On("GetByTelegramID", ctx, telegramID).Return(&domain.Resident{
			ID:          2,
			ApartmentID: 20,
			TelegramID:  telegramID,
			CarPlate:    nil,
			Status:      "active",
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(20)).Return(&domain.Apartment{
			ID:         20,
			BuildingID: buildingID,
			Number:     "15",
		}, nil)

		result, err := svc.ValidateResidentPersonalPass(ctx, token, 0, &buildingID)

		assert.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, "15", result.Apartment)
		assert.Equal(t, "", result.CarPlate)
	})

	t.Run("token with leading/trailing whitespace is valid", func(t *testing.T) {
		svc := newService()
		telegramID := int64(111222333)
		token := "  " + svc.GenerateResidentPersonalPassToken(telegramID) + "\n"

		passRepo2 := new(MockPassRepo)
		apartmentRepo2 := new(MockApartmentRepo)
		residentRepo2 := new(MockResidentRepo)
		ruleRepo2 := new(MockRuleRepo)
		scanEventRepo2 := new(MockScanEventRepo)
		svc2 := NewPassService(passRepo2, apartmentRepo2, residentRepo2, ruleRepo2, scanEventRepo2, secret, logger, noopMetrics)

		buildingID := int64(1)
		residentRepo2.On("GetByTelegramID", ctx, telegramID).Return(&domain.Resident{
			ID: 3, ApartmentID: 30, TelegramID: telegramID, Status: "active",
		}, nil)
		apartmentRepo2.On("GetByID", ctx, int64(30)).Return(&domain.Apartment{
			ID: 30, BuildingID: buildingID, Number: "7",
		}, nil)

		result, err := svc2.ValidateResidentPersonalPass(ctx, token, 0, &buildingID)

		assert.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("invalid prefix returns INVALID_PERSONAL_PASS", func(t *testing.T) {
		svc := newService()
		result, err := svc.ValidateResidentPersonalPass(ctx, "yardpass://pass/some-uuid", 0, nil)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, "INVALID_PERSONAL_PASS", result.Reason)
	})

	t.Run("tampered HMAC returns INVALID_PERSONAL_PASS", func(t *testing.T) {
		svc := newService()
		result, err := svc.ValidateResidentPersonalPass(ctx, "resident:123456789:tampered_hmac_value_here", 0, nil)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, "INVALID_PERSONAL_PASS", result.Reason)
	})

	t.Run("wrong secret returns INVALID_PERSONAL_PASS", func(t *testing.T) {
		svcGen := newService()
		telegramID := int64(555666777)
		token := svcGen.GenerateResidentPersonalPassToken(telegramID)

		svcVal := NewPassService(
			new(MockPassRepo), new(MockApartmentRepo), new(MockResidentRepo),
			new(MockRuleRepo), new(MockScanEventRepo),
			"different-secret", logger, noopMetrics,
		)
		result, err := svcVal.ValidateResidentPersonalPass(ctx, token, 0, nil)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, "INVALID_PERSONAL_PASS", result.Reason)
	})

	t.Run("building mismatch returns BUILDING_MISMATCH", func(t *testing.T) {
		passRepo := new(MockPassRepo)
		apartmentRepo := new(MockApartmentRepo)
		residentRepo := new(MockResidentRepo)
		ruleRepo := new(MockRuleRepo)
		scanEventRepo := new(MockScanEventRepo)
		svc := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, secret, logger, noopMetrics)

		telegramID := int64(444555666)
		token := svc.GenerateResidentPersonalPassToken(telegramID)

		guardBuildingID := int64(2)
		residentRepo.On("GetByTelegramID", ctx, telegramID).Return(&domain.Resident{
			ID: 4, ApartmentID: 40, TelegramID: telegramID, Status: "active",
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(40)).Return(&domain.Apartment{
			ID: 40, BuildingID: int64(1), Number: "99",
		}, nil)

		result, err := svc.ValidateResidentPersonalPass(ctx, token, 0, &guardBuildingID)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, "BUILDING_MISMATCH", result.Reason)
	})

	t.Run("inactive resident returns RESIDENT_NOT_FOUND", func(t *testing.T) {
		passRepo := new(MockPassRepo)
		apartmentRepo := new(MockApartmentRepo)
		residentRepo := new(MockResidentRepo)
		ruleRepo := new(MockRuleRepo)
		scanEventRepo := new(MockScanEventRepo)
		svc := NewPassService(passRepo, apartmentRepo, residentRepo, ruleRepo, scanEventRepo, secret, logger, noopMetrics)

		telegramID := int64(777888999)
		token := svc.GenerateResidentPersonalPassToken(telegramID)

		residentRepo.On("GetByTelegramID", ctx, telegramID).Return(&domain.Resident{
			ID: 5, ApartmentID: 50, TelegramID: telegramID, Status: "inactive",
		}, nil)

		result, err := svc.ValidateResidentPersonalPass(ctx, token, 0, nil)
		assert.NoError(t, err)
		assert.False(t, result.Valid)
		assert.Equal(t, "RESIDENT_NOT_FOUND", result.Reason)
	})
}

func TestNormalizeCarPlate(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"a123bc", "A123BC"},
		{"А123ВС77", "A123BC77"},
		{"а 123 в с 77", "A123BC77"},
		{"", ""},
		{"   ", ""},
		{"ЙЙЙ", ""},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.in), func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeCarPlate(tt.in))
		})
	}
}
