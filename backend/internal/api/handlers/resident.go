package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"yardpass/internal/domain"
	"yardpass/internal/errors"
	"yardpass/internal/service"
)

type ResidentHandler struct {
	residentService *service.ResidentService
}

func NewResidentHandler(residentService *service.ResidentService) *ResidentHandler {
	return &ResidentHandler{
		residentService: residentService,
	}
}

func (h *ResidentHandler) CreateResident(c *gin.Context) {
	var req service.CreateResidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	resident, err := h.residentService.CreateResident(c.Request.Context(), req)
	if err != nil {
		errors.BadRequest(c, "CREATE_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusCreated, resident)
}

func (h *ResidentHandler) BulkCreateResidents(c *gin.Context) {
	var requests []service.CreateResidentRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	residents, createErrors := h.residentService.BulkCreateResidents(c.Request.Context(), requests)

	response := gin.H{
		"created": len(residents),
	}

	if len(residents) > 0 {
		response["residents"] = residents
	}

	if len(createErrors) > 0 {
		response["errors"] = createErrors
	}

	c.JSON(http.StatusOK, response)
}

func (h *ResidentHandler) DeleteResident(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errors.BadRequest(c, "INVALID_ID", "Invalid resident ID format")
		return
	}

	if err := h.residentService.DeleteResident(c.Request.Context(), id); err != nil {
		errors.BadRequest(c, "DELETE_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Resident deleted successfully",
	})
}

func (h *ResidentHandler) ImportFromCSV(c *gin.Context) {
	buildingIDStr := c.Query("building_id")
	if buildingIDStr == "" {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id query parameter is required")
		return
	}

	buildingID, err := strconv.ParseInt(buildingIDStr, 10, 64)
	if err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "Invalid building ID format")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		errors.BadRequest(c, "MISSING_FILE", "file form field is required")
		return
	}

	src, err := file.Open()
	if err != nil {
		errors.BadRequest(c, "FILE_OPEN_ERROR", err.Error())
		return
	}
	defer src.Close()

	created, importErrors := h.residentService.ImportFromCSV(c.Request.Context(), src, buildingID)

	response := gin.H{
		"created": created,
	}

	if len(importErrors) > 0 {
		response["errors"] = importErrors
	}

	c.JSON(http.StatusOK, response)
}

func (h *ResidentHandler) ListResidents(c *gin.Context) {
	var filters domain.ResidentFilters

	if apartmentIDStr := c.Query("apartment_id"); apartmentIDStr != "" {
		if id, err := strconv.ParseInt(apartmentIDStr, 10, 64); err == nil {
			filters.ApartmentID = &id
		}
	}

	if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
		if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
			filters.BuildingID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		filters.Status = &status
	}

	filters.Limit = 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if id, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			filters.Limit = int(id)
		}
	}

	residents, err := h.residentService.ListResidents(c.Request.Context(), filters)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"residents": residents,
	})
}

