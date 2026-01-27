package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"yardpass/internal/config"
	"yardpass/internal/observability/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-ID"
	maxBodyLogSize  = 4096
	maskedValue     = "***MASKED***"
)

type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func newResponseWriter(w gin.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(lgr *zap.Logger, cfg config.LogConfig) gin.HandlerFunc {
	maskHeadersSet := toSet(cfg.MaskHeaders)
	maskBodyFieldsSet := toSet(cfg.MaskBodyFields)

	return func(c *gin.Context) {
		if cfg.Disabled {
			c.Next()
			return
		}

		start := time.Now()

		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header(RequestIDHeader, requestID)

		reqLogger := lgr.With(
			zap.String("request_id", requestID),
			zap.String("component", "http"),
		)

		var requestBody string
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= maxBodyLogSize {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		maskedHeaders := maskHeaders(c.Request.Header, maskHeadersSet)
		maskedBody := maskBodyFields(requestBody, maskBodyFieldsSet)

		reqLogger.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.Any("headers", maskedHeaders),
			zap.String("body", maskedBody),
		)

		ctx := logger.ToContext(c.Request.Context(), reqLogger.Sugar())
		c.Request = c.Request.WithContext(ctx)

		c.Set("request_id", requestID)

		rw := newResponseWriter(c.Writer)
		c.Writer = rw

		c.Next()

		duration := time.Since(start)

		var errMessages []string
		for _, err := range c.Errors {
			errMessages = append(errMessages, err.Error())
		}

		logFields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", duration),
			zap.Int("response_size", rw.body.Len()),
		}

		if len(errMessages) > 0 {
			logFields = append(logFields, zap.Strings("errors", errMessages))
		}

		if userID, exists := c.Get("user_id"); exists {
			logFields = append(logFields, zap.Any("user_id", userID))
		}

		switch {
		case rw.statusCode >= http.StatusInternalServerError:
			reqLogger.Error("Request completed with server error", logFields...)
		case rw.statusCode >= http.StatusBadRequest:
			reqLogger.Warn("Request completed with client error", logFields...)
		default:
			reqLogger.Info("Request completed", logFields...)
		}
	}
}

func toSet(slice []string) map[string]struct{} {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[strings.ToLower(s)] = struct{}{}
	}
	return set
}

func maskHeaders(headers http.Header, maskSet map[string]struct{}) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if _, shouldMask := maskSet[strings.ToLower(key)]; shouldMask {
			result[key] = maskedValue
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

func maskBodyFields(body string, maskSet map[string]struct{}) string {
	if body == "" {
		return ""
	}

	if len(body) > maxBodyLogSize {
		body = body[:maxBodyLogSize]
	}

	if len(maskSet) == 0 {
		return body
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	maskMapFields(data, maskSet)

	masked, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(masked)
}

func maskMapFields(data map[string]any, maskSet map[string]struct{}) {
	for key, value := range data {
		if _, shouldMask := maskSet[strings.ToLower(key)]; shouldMask {
			data[key] = maskedValue
			continue
		}

		switch v := value.(type) {
		case map[string]any:
			maskMapFields(v, maskSet)
		case []any:
			maskSliceFields(v, maskSet)
		}
	}
}

func maskSliceFields(slice []any, maskSet map[string]struct{}) {
	for _, item := range slice {
		if m, ok := item.(map[string]any); ok {
			maskMapFields(m, maskSet)
		}
	}
}
