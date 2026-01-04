package handlers

import (
	"fmt"
	"net/http"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	ruleRepo domain.RuleRepository
}

func NewRuleHandler(ruleRepo domain.RuleRepository) *RuleHandler {
	return &RuleHandler{
		ruleRepo: ruleRepo,
	}
}

type UpdateRuleRequest struct {
	QuietHoursStart            *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd              *string `json:"quiet_hours_end,omitempty"`
	DailyPassLimitPerApartment *int    `json:"daily_pass_limit_per_apartment,omitempty"`
	MaxPassDurationHours       *int    `json:"max_pass_duration_hours,omitempty"`
}

func (h *RuleHandler) Get(c *gin.Context) {
	buildingIDStr := c.Query("building_id")
	if buildingIDStr == "" {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id query parameter is required")
		return
	}

	var buildingID int64
	_, err := fmt.Sscanf(buildingIDStr, "%d", &buildingID)
	if err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "Invalid building ID format")
		return
	}

	rule, err := h.ruleRepo.GetByBuildingID(c.Request.Context(), buildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	if rule == nil {
		errors.NotFound(c, "RULE_NOT_FOUND", "Rules not found for this building")
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Update(c *gin.Context) {
	buildingIDStr := c.Query("building_id")
	if buildingIDStr == "" {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "building_id query parameter is required")
		return
	}

	var buildingID int64
	_, err := fmt.Sscanf(buildingIDStr, "%d", &buildingID)
	if err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "Invalid building ID format")
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	rule, err := h.ruleRepo.GetByBuildingID(c.Request.Context(), buildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	if rule == nil {
		rule = &domain.Rule{
			BuildingID:                 buildingID,
			DailyPassLimitPerApartment: 5,
			MaxPassDurationHours:       24,
		}
	}

	if req.QuietHoursStart != nil {
		rule.QuietHoursStart = req.QuietHoursStart
	}
	if req.QuietHoursEnd != nil {
		rule.QuietHoursEnd = req.QuietHoursEnd
	}
	if req.DailyPassLimitPerApartment != nil {
		rule.DailyPassLimitPerApartment = *req.DailyPassLimitPerApartment
	}
	if req.MaxPassDurationHours != nil {
		rule.MaxPassDurationHours = *req.MaxPassDurationHours
	}

	if rule.ID == 0 {
		err = h.ruleRepo.Create(c.Request.Context(), rule)
	} else {
		err = h.ruleRepo.Update(c.Request.Context(), rule)
	}

	if err != nil {
		errors.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, rule)
}
