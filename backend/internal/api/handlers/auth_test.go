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

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("login success", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc)

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
		assert.Equal(t, "access-tok", resp["access_token"])
		assert.Equal(t, "refresh-tok", resp["refresh_token"])
		svc.AssertExpectations(t)
	})

	t.Run("login invalid credentials", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc)

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
		h := NewAuthHandlerWithService(svc)

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

	t.Run("refresh success", func(t *testing.T) {
		svc := new(mockAuthService)
		h := NewAuthHandlerWithService(svc)

		r := gin.New()
		r.POST("/refresh", h.Refresh)

		svc.On("RefreshToken", mock.Anything, "valid-refresh-tok").Return(&domain.AuthTokens{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			ExpiresIn:    900,
		}, nil)

		body, _ := json.Marshal(RefreshRequest{RefreshToken: "valid-refresh-tok"})
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "new-access", resp["access_token"])
		svc.AssertExpectations(t)
	})
}
