package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"yardpass/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRuleRepo struct {
	mock.Mock
}

func (m *mockRuleRepo) GetByBuildingID(ctx context.Context, buildingID int64) (*domain.Rule, error) {
	args := m.Called(ctx, buildingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Rule), args.Error(1)
}

func (m *mockRuleRepo) Create(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *mockRuleRepo) Update(ctx context.Context, rule *domain.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func TestRuleHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get rule success", func(t *testing.T) {
		repo := new(mockRuleRepo)
		h := NewRuleHandler(repo)

		r := gin.New()
		r.GET("/rules", h.Get)

		rule := &domain.Rule{ID: 1, BuildingID: 1, DailyPassLimitPerApartment: 5}
		repo.On("GetByBuildingID", mock.Anything, int64(1)).Return(rule, nil)

		req := httptest.NewRequest(http.MethodGet, "/rules?building_id=1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})

	t.Run("missing building_id", func(t *testing.T) {
		repo := new(mockRuleRepo)
		h := NewRuleHandler(repo)

		r := gin.New()
		r.GET("/rules", h.Get)

		req := httptest.NewRequest(http.MethodGet, "/rules", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		repo.AssertNotCalled(t, "GetByBuildingID")
	})

	t.Run("rule not found", func(t *testing.T) {
		repo := new(mockRuleRepo)
		h := NewRuleHandler(repo)

		r := gin.New()
		r.GET("/rules", h.Get)

		repo.On("GetByBuildingID", mock.Anything, int64(99)).Return(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/rules?building_id=99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		repo.AssertExpectations(t)
	})
}

func TestRuleHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("update rule success", func(t *testing.T) {
		repo := new(mockRuleRepo)
		h := NewRuleHandler(repo)

		r := gin.New()
		r.PUT("/rules", h.Update)

		limit := 10
		rule := &domain.Rule{ID: 1, BuildingID: 1, DailyPassLimitPerApartment: 5}
		repo.On("GetByBuildingID", mock.Anything, int64(1)).Return(rule, nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Rule")).Return(nil)

		body, _ := json.Marshal(UpdateRuleRequest{DailyPassLimitPerApartment: &limit})
		req := httptest.NewRequest(http.MethodPut, "/rules?building_id=1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		repo.AssertExpectations(t)
	})
}
