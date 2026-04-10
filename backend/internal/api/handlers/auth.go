package handlers

import (
	"net/http"

	"yardpass/internal/config"
	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService domain.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// NewAuthHandlerWithService is an alias for NewAuthHandler for tests.
func NewAuthHandlerWithService(authService domain.AuthService, cfg *config.Config) *AuthHandler {
	return NewAuthHandler(authService, cfg)
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		errors.Unauthorized(c, "INVALID_CREDENTIALS", err.Error())
		return
	}

	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"expires_in": tokens.ExpiresIn,
		"token_type": "Bearer",
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(cookieRefreshToken)
	if err != nil || refreshToken == "" {
		errors.BadRequest(c, "INVALID_REQUEST", "refresh_token is required")
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		errors.Unauthorized(c, "INVALID_REFRESH_TOKEN", err.Error())
		return
	}

	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"expires_in": tokens.ExpiresIn,
		"token_type": "Bearer",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	response := gin.H{
		"user_id": userID,
		"role":    role,
	}

	// Include building_id if present in context
	if buildingID, exists := c.Get("building_id"); exists {
		response["building_id"] = buildingID
	}

	c.JSON(http.StatusOK, response)
}
