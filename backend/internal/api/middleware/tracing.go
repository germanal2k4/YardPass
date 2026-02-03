package middleware

import (
	"errors"

	"yardpass/internal/observability/tracer"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TracingMiddleware(t *tracer.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		fullServiceName := "yardpass"

		ctx, span := t.StartSpan(c.Request.Context(),
			tracer.TraceSafeString(fullServiceName+"::"+c.Request.URL.Path),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		errs := make([]error, 0, len(c.Errors))
		for _, err := range c.Errors {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			err := errors.Join(errs...)
			span.RecordError(err)
			span.SetStatus(codes.Error, tracer.TraceSafeString(err.Error()))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}
