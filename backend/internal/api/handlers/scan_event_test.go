package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yardpass/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockScanEventRepo struct {
	mock.Mock
}

func (m *mockScanEventRepo) Create(ctx context.Context, event *domain.ScanEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockScanEventRepo) List(ctx context.Context, filters domain.ScanEventFilters) ([]domain.ScanEvent, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]domain.ScanEvent), args.Error(1)
}

func (m *mockScanEventRepo) CountValidScansToday(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockScanEventRepo) GetEventsWithDetails(ctx context.Context, filters domain.ScanEventFilters, buildingID *int64) ([]domain.ScanEventWithDetails, error) {
	args := m.Called(ctx, filters, buildingID)
	return args.Get(0).([]domain.ScanEventWithDetails), args.Error(1)
}

func (m *mockScanEventRepo) GetStatistics(ctx context.Context, from, to *time.Time, buildingID *int64) (*domain.Statistics, error) {
	args := m.Called(ctx, from, to, buildingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Statistics), args.Error(1)
}

func TestScanEventHandler_ListEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list events success", func(t *testing.T) {
		repo := new(mockScanEventRepo)
		h := NewScanEventHandler(repo)

		r := gin.New()
		r.GET("/events", func(c *gin.Context) {
			c.Set("role", "guard")
			c.Set("building_id", int64(1))
			h.ListEvents(c)
		})

		events := []domain.ScanEventWithDetails{
			{ID: 1, PassID: uuid.New(), GuardUserID: 1, ScannedAt: time.Now(), Result: "valid", ApartmentNumber: "101"},
		}
		repo.On("GetEventsWithDetails", mock.Anything, mock.Anything, mock.Anything).Return(events, nil)

		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})

	t.Run("guard without building_id is rejected", func(t *testing.T) {
		repo := new(mockScanEventRepo)
		h := NewScanEventHandler(repo)

		r := gin.New()
		r.GET("/events", func(c *gin.Context) {
			c.Set("role", "guard")
			h.ListEvents(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		repo.AssertNotCalled(t, "GetEventsWithDetails")
	})
}
