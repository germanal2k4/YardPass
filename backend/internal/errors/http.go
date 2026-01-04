package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ErrorResponseJSON(c *gin.Context, code int, errorCode, message string) {
	c.JSON(code, ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	})
}

func BadRequest(c *gin.Context, errorCode, message string) {
	ErrorResponseJSON(c, http.StatusBadRequest, errorCode, message)
}

func Unauthorized(c *gin.Context, errorCode, message string) {
	ErrorResponseJSON(c, http.StatusUnauthorized, errorCode, message)
}

func Forbidden(c *gin.Context, errorCode, message string) {
	ErrorResponseJSON(c, http.StatusForbidden, errorCode, message)
}

func NotFound(c *gin.Context, errorCode, message string) {
	ErrorResponseJSON(c, http.StatusNotFound, errorCode, message)
}

func InternalServerError(c *gin.Context, errorCode, message string) {
	ErrorResponseJSON(c, http.StatusInternalServerError, errorCode, message)
}

