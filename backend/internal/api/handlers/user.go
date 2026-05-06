package handlers

import (
	"fmt"
	"net/http"
	"strconv"

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
		errors.BadRequestInvalidJSON(c)
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
	role, _ := c.Get("role")
	buildingIDCtx, hasBuildingID := c.Get("building_id")

	if role := c.Query("role"); role != "" {
		filters.Role = &role
	}

	if role == "admin" {
		if hasBuildingID {
			if id, ok := buildingIDCtx.(int64); ok {
				filters.BuildingID = &id
			}
		}
	} else if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
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
		errors.InternalServerError(c, "FETCH_FAILED", errors.UserMsgFetchFailed)
		return
	}

	for _, user := range users {
		user.PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (h *UserHandler) UpdateUserCredentials(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		errors.BadRequest(c, "INVALID_ID", "Некорректный идентификатор пользователя.")
		return
	}

	var req domain.UpdateUserCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequestInvalidJSON(c)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		errors.Unauthorized(c, "UNAUTHORIZED", "Требуется авторизация.")
		return
	}
	actorID, ok := userID.(int64)
	if !ok {
		errors.InternalServerError(c, "INVALID_ID", "Некорректный идентификатор пользователя в сессии.")
		return
	}

	updatedUser, updateErr := h.userService.UpdateCredentials(c.Request.Context(), actorID, targetID, req.Username, req.Password)
	if updateErr != nil {
		errors.BadRequest(c, "UPDATE_FAILED", updateErr.Error())
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}
