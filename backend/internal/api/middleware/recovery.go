package middleware

import (
	"runtime/debug"

	"yardpass/internal/errors"
	"yardpass/internal/observability/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RecoveryMiddleware(lgr *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				var reqLogger *zap.Logger
				if ctxLogger := logger.FromContext(c.Request.Context()); ctxLogger != nil {
					reqLogger = ctxLogger
				} else {
					reqLogger = lgr
				}

				requestID, _ := c.Get("request_id")

				reqLogger.Error("Panic recovered",
					zap.Any("panic", err),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("request_id", requestID),
				)

				c.Abort()

				errors.InternalServerError(c, "INTERNAL_ERROR", "Internal server error")
			}
		}()

		c.Next()
	}
}
