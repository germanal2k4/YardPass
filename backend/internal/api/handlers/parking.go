package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"yardpass/internal/domain"
	"yardpass/internal/errors"
)

type ParkingHandler struct {
	passService domain.PassService
}

func NewParkingHandler(passService domain.PassService) *ParkingHandler {
	return &ParkingHandler{
		passService: passService,
	}
}

func (h *ParkingHandler) GetOccupancy(c *gin.Context) {
	role, _ := c.Get("role")
	buildingID, _ := c.Get("building_id")

	var bID *int64
	if role == "guard" && buildingID != nil {
		id := buildingID.(int64)
		bID = &id
	} else if role == "admin" {
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		} else if buildingID != nil {
			id := buildingID.(int64)
			bID = &id
		}
	}

	if bID == nil {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id is required")
		return
	}

	activePasses, err := h.passService.GetActivePassesByBuilding(c.Request.Context(), *bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	occupied := len(activePasses)
	total := 100

	c.JSON(http.StatusOK, gin.H{
		"occupied": occupied,
		"total":    total,
		"free":     total - occupied,
		"percent":  float64(occupied) / float64(total) * 100,
	})
}

func (h *ParkingHandler) GetVehicles(c *gin.Context) {
	role, _ := c.Get("role")
	buildingID, _ := c.Get("building_id")

	var bID *int64
	if role == "guard" && buildingID != nil {
		id := buildingID.(int64)
		bID = &id
	} else if role == "admin" {
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		} else if buildingID != nil {
			id := buildingID.(int64)
			bID = &id
		}
	}

	if bID == nil {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id is required")
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	activePasses, err := h.passService.GetActivePassesByBuilding(c.Request.Context(), *bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	start := offset
	end := offset + limit
	if start > len(activePasses) {
		start = len(activePasses)
	}
	if end > len(activePasses) {
		end = len(activePasses)
	}

	vehicles := activePasses[start:end]

	c.JSON(http.StatusOK, gin.H{
		"vehicles": vehicles,
		"total":    len(activePasses),
		"limit":    limit,
		"offset":   offset,
	})
}
