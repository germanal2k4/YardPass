package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (*domain.AuthTokens, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthTokens), args.Error(1)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthTokens), args.Error(1)
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func testAuthConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour,
		},
		Cookie: config.CookieConfig{Secure: false, SameSite: "Lax"},
	}
}

type mockSubscriptionService struct {
	mock.Mock
}

func (m *mockSubscriptionService) Purchase(ctx context.Context, req domain.PurchaseSubscriptionRequest) (*domain.PurchaseSubscriptionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PurchaseSubscriptionResponse), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("login success", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc, testAuthConfig())

		r := gin.New()
		r.POST("/login", h.Login)

		svc.On("Login", mock.Anything, "guard1", "password123").Return(&domain.AuthTokens{
			AccessToken:  "access-tok",
			RefreshToken: "refresh-tok",
			ExpiresIn:    900,
		}, nil)

		body, _ := json.Marshal(LoginRequest{Username: "guard1", Password: "password123"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(900), resp["expires_in"])
		assert.Nil(t, resp["access_token"])
		cookies := w.Result().Cookies()
		var names []string
		for _, ck := range cookies {
			names = append(names, ck.Name)
		}
		assert.Contains(t, names, "access_token")
		assert.Contains(t, names, "refresh_token")
		svc.AssertExpectations(t)
	})

	t.Run("login invalid credentials", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc, testAuthConfig())

		r := gin.New()
		r.POST("/login", h.Login)

		svc.On("Login", mock.Anything, "baduser", "wrong").Return((*domain.AuthTokens)(nil), context.DeadlineExceeded)

		body, _ := json.Marshal(LoginRequest{Username: "baduser", Password: "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("login invalid request body", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc, testAuthConfig())

		r := gin.New()
		r.POST("/login", h.Login)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		svc.AssertNotCalled(t, "Login")
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("refresh success with cookie", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc, testAuthConfig())

		r := gin.New()
		r.POST("/refresh", h.Refresh)

		svc.On("RefreshToken", mock.Anything, "cookie-refresh").Return(&domain.AuthTokens{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			ExpiresIn:    900,
		}, nil)

		req := httptest.NewRequest(http.MethodPost, "/refresh", http.NoBody)
		req.Header.Set("Cookie", "refresh_token=cookie-refresh")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestAuthHandler_PurchaseSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("purchase success", func(t *testing.T) {
		authSvc := new(mockAuthService)
		subSvc := new(mockSubscriptionService)
		h := NewAuthHandlerWithServices(authSvc, subSvc, testAuthConfig())

		r := gin.New()
		r.POST("/purchase-subscription", h.PurchaseSubscription)

		reqBody := domain.PurchaseSubscriptionRequest{
			Email:          "client@example.com",
			BuildingName:   "ЖК Тестовый",
			ApartmentCount: 120,
			CardNumber:     "4111111111111111",
			CardHolder:     "IVAN PETROV",
			Expiry:         "12/30",
			CVV:            "123",
		}

		subSvc.On("Purchase", mock.Anything, reqBody).Return(&domain.PurchaseSubscriptionResponse{
			BuildingID:      10,
			BuildingName:    "ЖК Тестовый",
			ApartmentCount:  120,
			SubscriptionFee: 200000,
			Period:          "1 year",
			Email:           "client@example.com",
			Accounts: []domain.AccountCredentials{
				{Username: "admin_test_1234", Password: "secretA"},
				{Username: "guard_test_1234", Password: "secretG"},
			},
			Message: "ok",
		}, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/purchase-subscription", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		subSvc.AssertExpectations(t)
	})
}
