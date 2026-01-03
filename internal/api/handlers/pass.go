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
	CarPlate    string    `json:"car_plate" binding:"required"`
	GuestName   *string   `json:"guest_name,omitempty"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to" binding:"required"`
}

type ValidatePassRequest struct {
	QRUUID string `json:"qr_uuid" binding:"required"`
}

func (h *PassHandler) Create(c *gin.Context) {
	var req CreatePassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if req.ValidFrom.IsZero() {
		req.ValidFrom = time.Now()
	}

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

	passID, err := uuid.Parse(req.QRUUID)
	if err != nil {
		errors.BadRequest(c, "INVALID_QR_UUID", "Invalid QR code format")
		return
	}

	userID, _ := c.Get("user_id")
	var guardUserID int64
	if userID != nil {
		guardUserID = userID.(int64)
	}

	result, err := h.passService.ValidatePass(c.Request.Context(), passID, guardUserID)
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
	apartmentIDStr := c.Query("apartment_id")
	if apartmentIDStr == "" {
		errors.BadRequest(c, "MISSING_APARTMENT_ID", "apartment_id query parameter is required")
		return
	}

	apartmentID, err := strconv.ParseInt(apartmentIDStr, 10, 64)
	if err != nil {
		errors.BadRequest(c, "INVALID_APARTMENT_ID", "Invalid apartment ID format")
		return
	}

	passes, err := h.passService.GetActivePasses(c.Request.Context(), apartmentID)
	if err != nil {
		errors.InternalServerError(c, "FETCH_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passes": passes,
	})
}
