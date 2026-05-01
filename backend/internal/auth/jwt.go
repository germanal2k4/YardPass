package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/domain"
	"yardpass/internal/observability/metrics"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/bcrypt"
)

type JWTService struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	userRepo   domain.UserRepository
	authOps    *prometheus.CounterVec
}

func NewJWTService(cfg config.JWTConfig, userRepo domain.UserRepository, m *metrics.Metrics) *JWTService {
	authOps := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_auth",
			Name:      "operations_total",
			Help:      "Total number of authentication operations",
		},
		[]string{"operation", "result"},
	)

	m.GetRegistry().MustRegister(authOps)

	return &JWTService{
		secret:     cfg.Secret,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		userRepo:   userRepo,
		authOps:    authOps,
	}
}

type Claims struct {
	UserID     int64  `json:"user_id"`
	Role       string `json:"role"`
	BuildingID *int64 `json:"building_id,omitempty"`
	Type       string `json:"type"`
	jwt.RegisteredClaims
}

func (s *JWTService) Login(ctx context.Context, username, password string) (*domain.AuthTokens, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user == nil {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, errors.New("invalid credentials")
	}

	if user.Status != "active" {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, errors.New("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.generateToken(user.ID, user.Role, user.BuildingID, "access", s.accessTTL)
	if err != nil {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(user.ID, user.Role, user.BuildingID, "refresh", s.refreshTTL)
	if err != nil {
		s.authOps.WithLabelValues("login", "failure").Inc()
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	s.authOps.WithLabelValues("login", "success").Inc()

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *JWTService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	claims, err := s.validateToken(refreshToken)
	if err != nil {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != "refresh" {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, errors.New("token is not a refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, errors.New("user not found")
	}

	if user.Status != "active" {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, errors.New("user account is inactive")
	}

	accessToken, err := s.generateToken(user.ID, user.Role, user.BuildingID, "access", s.accessTTL)
	if err != nil {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.generateToken(user.ID, user.Role, user.BuildingID, "refresh", s.refreshTTL)
	if err != nil {
		s.authOps.WithLabelValues("refresh", "failure").Inc()
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	s.authOps.WithLabelValues("refresh", "success").Inc()

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *JWTService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	claims, err := s.validateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, errors.New("token is not an access token")
	}

	return &domain.TokenClaims{
		UserID:     claims.UserID,
		Role:       claims.Role,
		BuildingID: claims.BuildingID,
		Type:       claims.Type,
	}, nil
}

func (s *JWTService) generateToken(userID int64, role string, buildingID *int64, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:     userID,
		Role:       role,
		BuildingID: buildingID,
		Type:       tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
