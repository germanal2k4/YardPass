package handlers

import (
	"fmt"
	"net/http"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

type BuildingHandler struct {
	buildingRepo domain.BuildingRepository
}

func NewBuildingHandler(buildingRepo domain.BuildingRepository) *BuildingHandler {
	return &BuildingHandler{buildingRepo: buildingRepo}
}

type UpdateApartmentCountRequest struct {
	ApartmentCount int32 `json:"apartment_count" binding:"required"`
}

func (h *BuildingHandler) GetByID(c *gin.Context) {
	targetBuildingID, ok := resolveTargetBuildingID(c)
	if !ok {
		return
	}

	building, err := h.buildingRepo.GetByID(c.Request.Context(), targetBuildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_BUILDING_FAILED", err.Error())
		return
	}
	if building == nil {
		errors.NotFound(c, "BUILDING_NOT_FOUND", "building not found")
		return
	}

	c.JSON(http.StatusOK, building)
}

func resolveTargetBuildingID(c *gin.Context) (int64, bool) {
	role, _ := c.Get("role")
	buildingIDValue, hasBuildingID := c.Get("building_id")
	if role == "admin" && !hasBuildingID {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id is required for admin")
		return 0, false
	}

	if role == "admin" {
		return buildingIDValue.(int64), true
	}
	idParam := c.Param("id")
	if idParam == "" {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building id is required")
		return 0, false
	}
	var parsed int64
	if _, err := fmt.Sscanf(idParam, "%d", &parsed); err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "building id must be a number")
		return 0, false
	}
	return parsed, true
}

func (h *BuildingHandler) UpdateApartmentCount(c *gin.Context) {
	var req UpdateApartmentCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}
	if req.ApartmentCount <= 0 {
		errors.BadRequest(c, "INVALID_APARTMENT_COUNT", "apartment_count must be greater than zero")
		return
	}

	targetBuildingID, ok := resolveTargetBuildingID(c)
	if !ok {
		return
	}

	currentBuilding, err := h.buildingRepo.GetByID(c.Request.Context(), targetBuildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_BUILDING_FAILED", err.Error())
		return
	}
	if currentBuilding == nil {
		errors.NotFound(c, "BUILDING_NOT_FOUND", "building not found")
		return
	}
	if req.ApartmentCount < currentBuilding.ApartmentCount {
		errors.BadRequest(c, "INVALID_APARTMENT_COUNT", "apartment_count cannot be decreased")
		return
	}

	updated, err := h.buildingRepo.UpdateApartmentCount(c.Request.Context(), targetBuildingID, req.ApartmentCount)
	if err != nil {
		errors.InternalServerError(c, "UPDATE_BUILDING_FAILED", err.Error())
		return
	}
	if updated == nil {
		errors.NotFound(c, "BUILDING_NOT_FOUND", "building not found")
		return
	}

	c.JSON(http.StatusOK, updated)
}
