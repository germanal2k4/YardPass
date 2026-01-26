package handlers

import (
	"fmt"
	"net/http"

	"yardpass/internal/domain"
	"yardpass/internal/errors"
	"yardpass/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req domain.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	createdBy := userID.(int64)

	user, err := h.userService.RegisterUser(c.Request.Context(), req, createdBy)
	if err != nil {
		errors.BadRequest(c, "REGISTRATION_FAILED", err.Error())
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var filters domain.UserFilters

	if role := c.Query("role"); role != "" {
		filters.Role = &role
	}

	if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
		var buildingID int64
		if _, err := fmt.Sscanf(buildingIDStr, "%d", &buildingID); err == nil {
			filters.BuildingID = &buildingID
		}
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	filters.Limit = 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &filters.Limit); err != nil {
			filters.Limit = 100
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), filters)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	for _, user := range users {
		user.PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}
