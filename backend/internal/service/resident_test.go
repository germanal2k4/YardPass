package service

import (
	"context"
	"strings"
	"testing"

	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type mockResidentRepo struct{ mock.Mock }

func (m *mockResidentRepo) GetByID(ctx context.Context, id int64) (*domain.Resident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *mockResidentRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.Resident, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Resident), args.Error(1)
}

func (m *mockResidentRepo) Create(ctx context.Context, resident *domain.Resident) error {
	args := m.Called(ctx, resident)
	return args.Error(0)
}

func (m *mockResidentRepo) Update(ctx context.Context, resident *domain.Resident) error {
	args := m.Called(ctx, resident)
	return args.Error(0)
}

func (m *mockResidentRepo) SetCarPlate(ctx context.Context, id int64, carPlate *string) error {
	args := m.Called(ctx, id, carPlate)
	return args.Error(0)
}

func (m *mockResidentRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockResidentRepo) BulkCreate(ctx context.Context, residents []domain.Resident) error {
	args := m.Called(ctx, residents)
	return args.Error(0)
}

func (m *mockResidentRepo) List(ctx context.Context, filters domain.ResidentFilters) ([]domain.Resident, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.Resident), args.Error(1)
}

func TestResidentService_CreateResident(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}

	t.Run("create new resident", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentID := int64(1)
		name := "John"
		phone := "+79991234567"
		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{ID: apartmentID, BuildingID: 1}, nil)
		residentRepo.On("GetByTelegramID", ctx, int64(123)).Return(nil, nil)
		residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil)

		req := domain.CreateResidentRequest{
			ApartmentID: &apartmentID,
			TelegramID:  123,
			Name:        &name,
			Phone:       &phone,
		}
		resident, err := svc.CreateResident(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resident)
		assert.Equal(t, int64(123), resident.TelegramID)
		residentRepo.AssertExpectations(t)
	})

	t.Run("duplicate telegram_id returns error", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentID := int64(2)
		existing := &domain.Resident{ID: 1, TelegramID: 456, ApartmentID: 1}
		apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{ID: apartmentID}, nil)
		residentRepo.On("GetByTelegramID", ctx, int64(456)).Return(existing, nil)

		req := domain.CreateResidentRequest{
			ApartmentID: &apartmentID,
			TelegramID:  456,
		}
		resident, err := svc.CreateResident(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resident)
		assert.Contains(t, err.Error(), "telegram_id already exists")
		residentRepo.AssertNotCalled(t, "Update")
		residentRepo.AssertNotCalled(t, "Create")
		residentRepo.AssertExpectations(t)
	})

	t.Run("apartment not found", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentRepo.On("GetByID", ctx, int64(999)).Return(nil, nil)

		req := domain.CreateResidentRequest{
			ApartmentID: func() *int64 { v := int64(999); return &v }(),
			TelegramID:  123,
		}
		resident, err := svc.CreateResident(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resident)
		assert.Contains(t, err.Error(), "apartment not found")
		residentRepo.AssertNotCalled(t, "Create")
	})
}

func TestResidentService_CreateResident_ByApartmentNumber(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}
	residentRepo := new(mockResidentRepo)
	apartmentRepo := new(MockApartmentRepo)
	svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
	ctx := context.Background()

	buildingID := int64(10)
	apartmentID := int64(2)
	apartmentNumber := "102"
	apartmentRepo.On("GetByBuildingID", ctx, buildingID).Return([]domain.Apartment{
		{ID: apartmentID, Number: apartmentNumber, BuildingID: buildingID},
	}, nil)
	apartmentRepo.On("GetByID", ctx, apartmentID).Return(&domain.Apartment{ID: apartmentID, BuildingID: buildingID}, nil)
	residentRepo.On("GetByTelegramID", ctx, int64(999)).Return(nil, nil)
	residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil)

	req := domain.CreateResidentRequest{
		ApartmentNumber: &apartmentNumber,
		BuildingID:      &buildingID,
		TelegramID:      999,
	}
	resident, err := svc.CreateResident(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resident)
	assert.Equal(t, apartmentID, resident.ApartmentID)
}

func TestResidentService_ImportFromCSV(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}

	t.Run("missing header columns", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		csv := "name,phone\n123,John"
		count, errs := svc.ImportFromCSV(ctx, strings.NewReader(csv), 1)
		assert.Equal(t, 0, count)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "missing required column")
	})

	t.Run("valid csv import", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentRepo.On("GetByBuildingID", ctx, int64(1)).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: 1},
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(1)).Return(&domain.Apartment{ID: 1, BuildingID: 1}, nil)
		residentRepo.On("GetByTelegramID", ctx, int64(111)).Return(nil, nil)
		residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil)

		csv := "apartment,telegram_id,name,phone\n101,111,Ivan,+79991111111"
		count, errs := svc.ImportFromCSV(ctx, strings.NewReader(csv), 1)
		assert.Equal(t, 1, count)
		assert.Len(t, errs, 0)
		residentRepo.AssertExpectations(t)
	})
}
