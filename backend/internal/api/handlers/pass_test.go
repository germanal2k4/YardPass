package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

type mockPassService struct {
	mock.Mock
}

func (m *mockPassService) CreatePass(ctx context.Context, req domain.CreatePassRequest) (*domain.Pass, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Pass), args.Error(1)
}

func (m *mockPassService) ValidatePass(ctx context.Context, passID uuid.UUID, guardUserID int64) (*domain.PassValidationResult, error) {
	args := m.Called(ctx, passID, guardUserID)
	return args.Get(0).(*domain.PassValidationResult), args.Error(1)
}

func (m *mockPassService) ValidatePassByCarPlate(ctx context.Context, carPlate string, guardUserID int64, buildingID *int64) (*domain.PassValidationResult, error) {
	args := m.Called(ctx, carPlate, guardUserID, buildingID)
	return args.Get(0).(*domain.PassValidationResult), args.Error(1)
}

func (m *mockPassService) RevokePass(ctx context.Context, passID uuid.UUID, revokedBy int64) error {
	args := m.Called(ctx, passID, revokedBy)
	return args.Error(0)
}

func (m *mockPassService) GetActivePasses(ctx context.Context, apartmentID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, apartmentID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *mockPassService) GetActivePassesByResident(ctx context.Context, residentID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, residentID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *mockPassService) GetActivePassesByBuilding(ctx context.Context, buildingID int64) ([]domain.Pass, error) {
	args := m.Called(ctx, buildingID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func (m *mockPassService) SearchPassesByCarPlate(ctx context.Context, carPlate string, buildingID *int64) ([]domain.Pass, error) {
	args := m.Called(ctx, carPlate, buildingID)
	return args.Get(0).([]domain.Pass), args.Error(1)
}

func TestPassHandler_Validate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid by qr_uuid", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/validate", func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("role", "guard")
			c.Set("building_id", int64(1))
			h.Validate(c)
		})

		passID := uuid.New()
		validTo := time.Now().Add(time.Hour)
		svc.On("ValidatePass", mock.Anything, passID, int64(1)).Return(&domain.PassValidationResult{
			Valid:     true,
			CarPlate:  "A123BC",
			Apartment: "101",
			ValidTo:   &validTo,
		}, nil)

		body, _ := json.Marshal(ValidatePassRequest{QRUUID: passID.String()})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["valid"].(bool))
		assert.Equal(t, "A123BC", resp["car_plate"])
		svc.AssertExpectations(t)
	})

	t.Run("invalid pass returns reason", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/validate", func(c *gin.Context) {
			c.Set("user_id", int64(1))
			h.Validate(c)
		})

		passID := uuid.New()
		svc.On("ValidatePass", mock.Anything, passID, int64(1)).Return(&domain.PassValidationResult{
			Valid:  false,
			Reason: "PASS_EXPIRED",
		}, nil)

		body, _ := json.Marshal(ValidatePassRequest{QRUUID: passID.String()})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["valid"].(bool))
		assert.Equal(t, "PASS_EXPIRED", resp["reason"])
		svc.AssertExpectations(t)
	})

	t.Run("missing qr_uuid and car_plate returns 400", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/validate", h.Validate)

		body, _ := json.Marshal(ValidatePassRequest{})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertNotCalled(t, "ValidatePass")
	})

	t.Run("valid by car_plate", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/validate", func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("building_id", int64(1))
			h.Validate(c)
		})

		bID := int64(1)
		validTo := time.Now().Add(time.Hour)
		svc.On("ValidatePassByCarPlate", mock.Anything, "A123BC", int64(1), &bID).Return(&domain.PassValidationResult{
			Valid:     true,
			CarPlate:  "A123BC",
			Apartment: "101",
			ValidTo:   &validTo,
		}, nil)

		body, _ := json.Marshal(ValidatePassRequest{CarPlate: "A123BC"})
		req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["valid"].(bool))
		svc.AssertExpectations(t)
	})
}

func TestPassHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create pass success", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/create", func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("resident_id", int64(1))
			h.Create(c)
		})

		carPlate := "A123BC"
		validTo := time.Now().Add(2 * time.Hour)
		pass := &domain.Pass{
			ID:          uuid.New(),
			ApartmentID: 1,
			CarPlate:    &carPlate,
			ValidTo:     validTo,
			Status:      "active",
		}
		svc.On("CreatePass", mock.Anything, mock.MatchedBy(func(r domain.CreatePassRequest) bool {
			return r.ApartmentID == 1 && r.CarPlate != nil && *r.CarPlate == carPlate
		})).Return(pass, nil)

		body, _ := json.Marshal(CreatePassRequest{
			ApartmentID: 1,
			CarPlate:    &carPlate,
			ValidTo:     validTo,
		})
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("create pass invalid request", func(t *testing.T) {
		svc := new(mockPassService)
		h := NewPassHandler(svc)

		r := gin.New()
		r.POST("/create", h.Create)

		body := []byte(`{"apartment_id": 1}`) // missing valid_to
		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertNotCalled(t, "CreatePass")
	})
}
