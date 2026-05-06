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

	if apartmentNumber := c.Query("apartment_number"); apartmentNumber != "" {
		filters.ApartmentNumber = &apartmentNumber
	}

	if carPlate := c.Query("car_plate"); carPlate != "" {
		filters.CarPlate = &carPlate
	}

	role, exists := c.Get("role")
	if !exists {
		errors.Unauthorized(c, "MISSING_ROLE", "Роль пользователя не найдена. Обратитесь к администратору.")
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		errors.InternalServerError(c, "INVALID_ROLE", "Некорректный тип роли в сессии.")
		return
	}

	buildingID, _ := c.Get("building_id")

	var bID *int64

	switch roleStr {
	case "superuser":
		if buildingIDStr := c.Query("building_id"); buildingIDStr != "" {
			if id, err := strconv.ParseInt(buildingIDStr, 10, 64); err == nil {
				bID = &id
			}
		}
	case "admin", "guard":
		if buildingID == nil {
			errors.Unauthorized(c, "MISSING_BUILDING_ID", "Для вашей роли требуется привязка к зданию (building_id).")
			return
		}
		id, ok := buildingID.(int64)
		if !ok {
			errors.InternalServerError(c, "INVALID_BUILDING_ID", "Некорректный building_id в сессии. Войдите снова.")
			return
		}
		bID = &id
	}

	events, err := h.scanEventRepo.GetEventsWithDetails(c.Request.Context(), filters, bID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", errors.UserMsgFetchFailed)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})
}
