package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Сообщения для пользователя (API, бот через те же сервисы).
const (
	UserMsgInvalidJSON       = "Некорректный формат данных. Проверьте, что все обязательные поля заполнены."
	UserMsgFetchFailed       = "Не удалось загрузить данные. Попробуйте позже."
	UserMsgSaveFailed        = "Не удалось сохранить данные. Попробуйте позже."
	UserMsgSearchFailed      = "Не удалось выполнить поиск. Попробуйте позже."
	UserMsgValidatePassFailed = "Не удалось проверить пропуск. Попробуйте ещё раз."
	UserMsgInternalError     = "Произошла внутренняя ошибка. Повторите попытку позже."
	UserMsgFileOpenFailed    = "Не удалось прочитать файл. Проверьте формат и повторите попытку."
	UserMsgExcelFailed       = "Не удалось сформировать отчёт. Попробуйте позже."
	UserMsgUpdateFailed      = "Не удалось обновить данные. Попробуйте позже."
	UserMsgDeleteFailed      = "Не удалось удалить запись. Попробуйте позже."
	UserMsgRevokeFailed      = "Не удалось отозвать пропуск. Попробуйте позже."
	UserMsgCreatePassFailed  = "Не удалось создать пропуск. Попробуйте позже."
	UserMsgBuildingFetch     = "Не удалось загрузить сведения о здании. Попробуйте позже."
	UserMsgBuildingUpdate    = "Не удалось обновить сведения о здании. Попробуйте позже."
)

// BadRequestInvalidJSON — 400 при ошибке разбора/валидации JSON без английских деталей валидатора.
func BadRequestInvalidJSON(c *gin.Context) {
	ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_REQUEST", UserMsgInvalidJSON)
}

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

