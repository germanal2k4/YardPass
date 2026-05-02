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
		errors.BadRequest(c, "MISSING_BUILDING_ID", "Укажите параметр запроса building_id.")
		return
	}

	var buildingID int64
	_, err := fmt.Sscanf(buildingIDStr, "%d", &buildingID)
	if err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "Некорректный формат building_id.")
		return
	}

	rule, err := h.ruleRepo.GetByBuildingID(c.Request.Context(), buildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", errors.UserMsgFetchFailed)
		return
	}

	if rule == nil {
		errors.NotFound(c, "RULE_NOT_FOUND", "Правила для этого здания не найдены.")
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Update(c *gin.Context) {
	buildingIDStr := c.Query("building_id")
	if buildingIDStr == "" {
		errors.BadRequest(c, "MISSING_BUILDING_ID", "Укажите параметр запроса building_id.")
		return
	}

	var buildingID int64
	_, err := fmt.Sscanf(buildingIDStr, "%d", &buildingID)
	if err != nil {
		errors.BadRequest(c, "INVALID_BUILDING_ID", "Некорректный формат building_id.")
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequestInvalidJSON(c)
		return
	}

	rule, err := h.ruleRepo.GetByBuildingID(c.Request.Context(), buildingID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", errors.UserMsgFetchFailed)
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
		rule.DailyPassLimitPerApartment = int32(*req.DailyPassLimitPerApartment)
	}
	if req.MaxPassDurationHours != nil {
		rule.MaxPassDurationHours = int32(*req.MaxPassDurationHours)
	}

	if rule.ID == 0 {
		err = h.ruleRepo.Create(c.Request.Context(), rule)
	} else {
		err = h.ruleRepo.Update(c.Request.Context(), rule)
	}

	if err != nil {
		errors.InternalServerError(c, "UPDATE_FAILED", errors.UserMsgUpdateFailed)
		return
	}

	c.JSON(http.StatusOK, rule)
}
