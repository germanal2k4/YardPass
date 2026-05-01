package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yardpass/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBuildingRepository struct {
	mock.Mock
}

func (m *mockBuildingRepository) GetByID(ctx context.Context, id int64) (*domain.Building, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Building), args.Error(1)
}

func (m *mockBuildingRepository) List(ctx context.Context) ([]domain.Building, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Building), args.Error(1)
}

func (m *mockBuildingRepository) Create(ctx context.Context, building *domain.Building) error {
	args := m.Called(ctx, building)
	return args.Error(0)
}

func (m *mockBuildingRepository) UpdateApartmentCount(ctx context.Context, id int64, apartmentCount int32) (*domain.Building, error) {
	args := m.Called(ctx, id, apartmentCount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Building), args.Error(1)
}

func TestBuildingHandler_UpdateApartmentCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects decreasing apartment count", func(t *testing.T) {
		repo := new(mockBuildingRepository)
		handler := NewBuildingHandler(repo)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", "admin")
			c.Set("building_id", int64(1))
			c.Next()
		})
		r.PUT("/api/v1/buildings/:id/apartment-count", handler.UpdateApartmentCount)

		repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Building{
			ID:             1,
			ApartmentCount: 100,
		}, nil)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/buildings/1/apartment-count", strings.NewReader(`{"apartment_count":90}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		repo.AssertNotCalled(t, "UpdateApartmentCount", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("accepts increasing apartment count", func(t *testing.T) {
		repo := new(mockBuildingRepository)
		handler := NewBuildingHandler(repo)

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", "admin")
			c.Set("building_id", int64(1))
			c.Next()
		})
		r.PUT("/api/v1/buildings/:id/apartment-count", handler.UpdateApartmentCount)

		repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Building{
			ID:             1,
			ApartmentCount: 100,
		}, nil)
		repo.On("UpdateApartmentCount", mock.Anything, int64(1), int32(120)).Return(&domain.Building{
			ID:             1,
			ApartmentCount: 120,
		}, nil)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/buildings/1/apartment-count", strings.NewReader(`{"apartment_count":120}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})
}
