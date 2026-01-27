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

				var reqLogger *zap.SugaredLogger
				if ctxLogger := logger.FromContext(c.Request.Context()); ctxLogger != nil {
					reqLogger = ctxLogger
				} else {
					reqLogger = lgr.Sugar()
				}

				requestID, _ := c.Get("request_id")

				reqLogger.Errorw("Panic recovered",
					"panic", err,
					"stack", string(stack),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
					"request_id", requestID,
				)

				c.Abort()

				errors.InternalServerError(c, "INTERNAL_ERROR", "Internal server error")
			}
		}()

		c.Next()
	}
}
