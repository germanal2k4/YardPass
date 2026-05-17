package service

import (
	"context"
	"strings"
	"testing"

	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *mockResidentRepo) ListExistingTelegramIDs(ctx context.Context, telegramIDs []int64) (map[int64]struct{}, error) {
	args := m.Called(ctx, telegramIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]struct{}), args.Error(1)
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

func (m *mockResidentRepo) SetTimezone(ctx context.Context, id int64, timezone *string) error {
	args := m.Called(ctx, id, timezone)
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

func (m *mockResidentRepo) ListActiveWithCarPlate(ctx context.Context, buildingID *int64) ([]domain.Resident, error) {
	args := m.Called(ctx, buildingID)
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
		assert.Contains(t, err.Error(), "Telegram ID уже")
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
		assert.Contains(t, err.Error(), "Квартира не найдена")
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
		assert.Contains(t, errs[0].Error(), "обязательная колонка")
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
		residentRepo.On("ListExistingTelegramIDs", ctx, mock.MatchedBy(func(ids []int64) bool {
			return len(ids) == 1 && ids[0] == 111
		})).Return(map[int64]struct{}{}, nil)
		residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil)

		csv := "apartment,telegram_id,name,phone\n101,111,Ivan,+79991111111"
		count, errs := svc.ImportFromCSV(ctx, strings.NewReader(csv), 1)
		assert.Equal(t, 1, count)
		assert.Len(t, errs, 0)
		residentRepo.AssertExpectations(t)
	})

	t.Run("csv normalizes mixed phone formats", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentRepo.On("GetByBuildingID", ctx, int64(1)).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: 1},
			{ID: 2, Number: "102", BuildingID: 1},
		}, nil)
		apartmentRepo.On("GetByID", ctx, mock.AnythingOfType("int64")).Return(&domain.Apartment{ID: 1, BuildingID: 1}, nil)
		residentRepo.On("ListExistingTelegramIDs", ctx, mock.Anything).Return(map[int64]struct{}{}, nil)

		capturedPhones := make([]string, 0, 2)
		residentRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.Resident) bool {
			if r.Phone != nil {
				capturedPhones = append(capturedPhones, *r.Phone)
			}
			return true
		})).Return(nil).Twice()

		csv := "apartment,telegram_id,name,phone\n" +
			"101,111,Ivan,8 (900) 123-45-67\n" +
			"102,222,Anna,+7 900 765 43 21"
		count, errs := svc.ImportFromCSV(ctx, strings.NewReader(csv), 1)
		assert.Equal(t, 2, count)
		assert.Empty(t, errs)
		assert.ElementsMatch(t, []string{"+79001234567", "+79007654321"}, capturedPhones)
	})

	t.Run("csv rejects unparseable phone", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)
		ctx := context.Background()

		apartmentRepo.On("GetByBuildingID", ctx, int64(1)).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: 1},
		}, nil)

		csv := "apartment,telegram_id,name,phone\n101,111,Ivan,12345"
		count, errs := svc.ImportFromCSV(ctx, strings.NewReader(csv), 1)
		assert.Equal(t, 0, count)
		require.NotEmpty(t, errs)
		assert.Contains(t, errs[0].Error(), "телефон")
		residentRepo.AssertNotCalled(t, "ListExistingTelegramIDs", mock.Anything, mock.Anything)
		residentRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}

func TestResidentService_BulkCreateResidents(t *testing.T) {
	logger := zap.NewNop()
	noopMetrics := &metrics.Metrics{}
	ctx := context.Background()

	t.Run("apartments loaded once across many rows", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)

		buildingID := int64(1)
		apartmentRepo.On("GetByBuildingID", ctx, buildingID).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: buildingID},
			{ID: 2, Number: "102", BuildingID: buildingID},
		}, nil).Once()
		apartmentRepo.On("GetByID", ctx, mock.AnythingOfType("int64")).
			Return(&domain.Apartment{ID: 1, BuildingID: buildingID}, nil)
		residentRepo.On("ListExistingTelegramIDs", ctx, mock.Anything).
			Return(map[int64]struct{}{}, nil).Once()
		residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil)

		apt1 := "101"
		apt2 := "102"
		reqs := []domain.CreateResidentRequest{
			{ApartmentNumber: &apt1, BuildingID: &buildingID, TelegramID: 1001},
			{ApartmentNumber: &apt2, BuildingID: &buildingID, TelegramID: 1002},
			{ApartmentNumber: &apt1, BuildingID: &buildingID, TelegramID: 1003},
		}
		residents, errs := svc.BulkCreateResidents(ctx, reqs)
		assert.Len(t, residents, 3)
		assert.Empty(t, errs)
		apartmentRepo.AssertNumberOfCalls(t, "GetByBuildingID", 1)
		residentRepo.AssertNumberOfCalls(t, "ListExistingTelegramIDs", 1)
	})

	t.Run("flags duplicates within the request and pre-existing telegram_ids", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)

		buildingID := int64(1)
		apartmentRepo.On("GetByBuildingID", ctx, buildingID).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: buildingID},
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(1)).Return(&domain.Apartment{ID: 1, BuildingID: buildingID}, nil)
		residentRepo.On("ListExistingTelegramIDs", ctx, mock.Anything).
			Return(map[int64]struct{}{500: {}}, nil)
		residentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Resident")).Return(nil).Once()

		apt := "101"
		reqs := []domain.CreateResidentRequest{
			{ApartmentNumber: &apt, BuildingID: &buildingID, TelegramID: 100},
			{ApartmentNumber: &apt, BuildingID: &buildingID, TelegramID: 100},
			{ApartmentNumber: &apt, BuildingID: &buildingID, TelegramID: 500},
		}
		residents, errs := svc.BulkCreateResidents(ctx, reqs)
		assert.Len(t, residents, 1)
		require.Len(t, errs, 2)
		assert.Equal(t, 2, errs[0].Row)
		assert.Contains(t, errs[0].Error, "Дубликат")
		assert.Equal(t, 3, errs[1].Row)
		assert.Contains(t, errs[1].Error, "Telegram ID")
	})

	t.Run("normalizes phone for every accepted row", func(t *testing.T) {
		residentRepo := new(mockResidentRepo)
		apartmentRepo := new(MockApartmentRepo)
		svc := NewResidentService(residentRepo, apartmentRepo, logger, noopMetrics)

		buildingID := int64(1)
		apartmentRepo.On("GetByBuildingID", ctx, buildingID).Return([]domain.Apartment{
			{ID: 1, Number: "101", BuildingID: buildingID},
		}, nil)
		apartmentRepo.On("GetByID", ctx, int64(1)).Return(&domain.Apartment{ID: 1, BuildingID: buildingID}, nil)
		residentRepo.On("ListExistingTelegramIDs", ctx, mock.Anything).
			Return(map[int64]struct{}{}, nil)

		var captured string
		residentRepo.On("Create", ctx, mock.MatchedBy(func(r *domain.Resident) bool {
			if r.Phone != nil {
				captured = *r.Phone
			}
			return true
		})).Return(nil).Once()

		apt := "101"
		rawPhone := "8 (900) 123-45-67"
		reqs := []domain.CreateResidentRequest{
			{ApartmentNumber: &apt, BuildingID: &buildingID, TelegramID: 1, Phone: &rawPhone},
		}
		residents, errs := svc.BulkCreateResidents(ctx, reqs)
		assert.Len(t, residents, 1)
		assert.Empty(t, errs)
		assert.Equal(t, "+79001234567", captured)
	})
}
