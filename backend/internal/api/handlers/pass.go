package handlers

import (
	"net/http"
	"strconv"
	"time"

	"yardpass/internal/domain"
	"yardpass/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PassHandler struct {
	passService domain.PassService
}

func NewPassHandler(passService domain.PassService) *PassHandler {
	return &PassHandler{
		passService: passService,
	}
}

type CreatePassRequest struct {
	ApartmentID int64     `json:"apartment_id" binding:"required"`
	CarPlate    *string   `json:"car_plate,omitempty"` // NULL for pedestrian guests
	GuestName   *string   `json:"guest_name,omitempty"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to" binding:"required"`
}

type ValidatePassRequest struct {
	QRUUID   string `json:"qr_uuid,omitempty"`   // UUID из QR кода (опционально)
	CarPlate string `json:"car_plate,omitempty"`  // Номер машины (опционально, альтернатива QR)
}

func (h *PassHandler) Create(c *gin.Context) {
	var req CreatePassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if req.ValidFrom.IsZero() {
		// Используем UTC для единообразия с БД
		req.ValidFrom = time.Now().UTC()
	} else {
		// Нормализуем к UTC, если время пришло с часовым поясом
		req.ValidFrom = req.ValidFrom.UTC()
	}
	
	// Нормализуем ValidTo к UTC
	req.ValidTo = req.ValidTo.UTC()

	createReq := domain.CreatePassRequest{
		ApartmentID: req.ApartmentID,
		CarPlate:    req.CarPlate,
		GuestName:   req.GuestName,
		ValidFrom:   req.ValidFrom,
		ValidTo:     req.ValidTo,
	}

	pass, err := h.passService.CreatePass(c.Request.Context(), createReq)
	if err != nil {
		errors.BadRequest(c, "CREATE_PASS_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusCreated, pass)
}

func (h *PassHandler) GetByID(c *gin.Context) {
	_ = c.Param("id")
	errors.ErrorResponseJSON(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Get pass by ID not yet implemented")
}

func (h *PassHandler) Revoke(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		errors.BadRequest(c, "INVALID_UUID", "Invalid pass ID format")
		return
	}

	userID, _ := c.Get("user_id")
	var revokedBy int64
	if userID != nil {
		revokedBy = userID.(int64)
	}

	if err := h.passService.RevokePass(c.Request.Context(), id, revokedBy); err != nil {
		errors.BadRequest(c, "REVOKE_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pass revoked successfully",
		"pass_id": id.String(),
	})
}

func (h *PassHandler) Validate(c *gin.Context) {
	var req ValidatePassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	var guardUserID int64
	if userID != nil {
		guardUserID = userID.(int64)
	}

	buildingID, _ := c.Get("building_id")
	var bID *int64
	if buildingID != nil {
		if id, ok := buildingID.(int64); ok {
			bID = &id
		}
	}

	var result *domain.PassValidationResult
	var err error

	// Валидация по номеру машины (приоритет, если указан)
	if req.CarPlate != "" {
		result, err = h.passService.ValidatePassByCarPlate(c.Request.Context(), req.CarPlate, guardUserID, bID)
	} else if req.QRUUID != "" {
		// Валидация по QR коду
		passID, parseErr := uuid.Parse(req.QRUUID)
		if parseErr != nil {
			errors.BadRequest(c, "INVALID_QR_UUID", "Invalid QR code format")
			return
		}
		result, err = h.passService.ValidatePass(c.Request.Context(), passID, guardUserID)
	} else {
		errors.BadRequest(c, "MISSING_PARAMETER", "Either qr_uuid or car_plate must be provided")
		return
	}

	if err != nil {
		errors.InternalServerError(c, "VALIDATION_ERROR", err.Error())
		return
	}

	if result.Valid {
		c.JSON(http.StatusOK, gin.H{
			"valid":     true,
			"car_plate": result.CarPlate,
			"apartment": result.Apartment,
			"valid_to":  result.ValidTo,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"reason": result.Reason,
		})
	}
}

func (h *PassHandler) GetActive(c *gin.Context) {
	role, _ := c.Get("role")
	buildingID, _ := c.Get("building_id")

	var passes []*domain.Pass
	var err error

	if role == "guard" && buildingID != nil {
		passes, err = h.passService.GetActivePassesByBuilding(c.Request.Context(), buildingID.(int64))
	} else if apartmentIDStr := c.Query("apartment_id"); apartmentIDStr != "" {
		apartmentID, parseErr := strconv.ParseInt(apartmentIDStr, 10, 64)
		if parseErr != nil {
			errors.BadRequest(c, "INVALID_APARTMENT_ID", "Invalid apartment ID format")
			return
		}
		passes, err = h.passService.GetActivePasses(c.Request.Context(), apartmentID)
	} else if role == "admin" && buildingID != nil {
		passes, err = h.passService.GetActivePassesByBuilding(c.Request.Context(), buildingID.(int64))
	} else {
		errors.BadRequest(c, "MISSING_PARAMETER", "apartment_id or building_id required")
		return
	}

	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passes": passes,
	})
}

func (h *PassHandler) Search(c *gin.Context) {
	carPlate := c.Query("car_plate")
	if carPlate == "" {
		errors.BadRequest(c, "MISSING_CAR_PLATE", "car_plate query parameter is required")
		return
	}

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
		}
	}

	passes, err := h.passService.SearchPassesByCarPlate(c.Request.Context(), carPlate, bID)
	if err != nil {
		errors.InternalServerError(c, "SEARCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passes": passes,
	})
}
