package handlers

import (
	"net/http"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService domain.AuthService
}

func NewAuthHandler(authService domain.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// NewAuthHandlerWithService is an alias for NewAuthHandler for tests.
func NewAuthHandlerWithService(authService domain.AuthService) *AuthHandler {
	return NewAuthHandler(authService)
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
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

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		errors.Unauthorized(c, "INVALID_REFRESH_TOKEN", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    "Bearer",
	})
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
