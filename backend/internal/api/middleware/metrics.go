package middleware

import (
	"strconv"
	"time"
	"yardpass/internal/observability/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func MetricsMiddleware(metrics *metrics.Metrics) gin.HandlerFunc {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_api",
			Name:      "http_requests_total",
			Help:      "Total number of requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "yardpass_api",
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	errorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "yardpass_api",
			Name:      "http_errors_total",
			Help:      "Total number of errors",
		},
		[]string{"method", "path", "status"},
	)

	registry := metrics.GetRegistry()

	registry.MustRegister(requestsTotal, requestDuration, errorsTotal)

	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		duration := time.Since(startTime)
		requestsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status())).Inc()
		requestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status())).Observe(duration.Seconds())
		errorsTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status())).Inc()
		c.Next()
	}
}
