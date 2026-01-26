package handlers

import (
	"net/http"
	"strconv"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

type ScanEventHandler struct {
	scanEventRepo domain.ScanEventRepository
}

func NewScanEventHandler(scanEventRepo domain.ScanEventRepository) *ScanEventHandler {
	return &ScanEventHandler{
		scanEventRepo: scanEventRepo,
	}
}

func (h *ScanEventHandler) ListEvents(c *gin.Context) {
	var filters domain.ScanEventFilters

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		} else {
			filters.Limit = 20
		}
	} else {
		filters.Limit = 20
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filters.From = &from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filters.To = &to
		}
	}

	if result := c.Query("result"); result != "" {
		filters.Result = &result
	}

	role, exists := c.Get("role")
	if !exists {
		errors.Unauthorized(c, "MISSING_ROLE", "User role not found")
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		errors.InternalServerError(c, "INVALID_ROLE", "Invalid role type")
		return
	}

	buildingID, _ := c.Get("building_id")

	var bID *int64

	if roleStr == "superuser" {
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		}
	} else if roleStr == "admin" || roleStr == "guard" {
		if buildingID != nil {
			if id, ok := buildingID.(int64); ok {
				bID = &id
			}
		}
	}

	events, err := h.scanEventRepo.GetEventsWithDetails(c.Request.Context(), filters, bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}
