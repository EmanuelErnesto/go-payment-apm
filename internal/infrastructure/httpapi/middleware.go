package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
	"github.com/EmanuelErnesto/go-payment-apm/internal/infrastructure/logger"
)

const maxRequestBodyBytes = 64 << 10

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Cache-Control", "no-store")
		c.Next()
	}
}

func limitBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		c.Next()
	}
}

// requestContext must run right after apmgin.Middleware: at that point the
// APM transaction is already in c.Request.Context(), so the request-scoped
// logger is created once with the trace ids and travels down through ctx.
func requestContext(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = logger.WithContext(ctx, logger.WithTrace(ctx, base))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// requestLogger deliberately logs only the route template, never the raw URL,
// query string or body, to keep user data out of the log pipeline.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := zap.InfoLevel
		if status >= http.StatusInternalServerError {
			level = zap.ErrorLevel
		}

		logger.FromContext(c.Request.Context()).Log(level, "http request",
			zap.String("http.request.method", c.Request.Method),
			zap.String("http.route", c.FullPath()),
			zap.Int("http.response.status_code", status),
			zap.Int64("event.duration", time.Since(start).Nanoseconds()),
		)
	}
}

// errorHandler centralises the translation of errors pushed with c.Error()
// into HTTP responses. Handlers stay free of logging and APM concerns.
//
// apmgin.Middleware reports every error left in c.Errors to APM after the
// chain finishes. Client errors are therefore removed from c.Errors here so
// that only real failures (5xx) show up as APM errors.
func errorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		last := c.Errors.Last().Err
		var serverErrors []*gin.Error
		for _, ginErr := range c.Errors {
			if !errors.Is(ginErr.Err, domain.ErrInvalidInput) {
				serverErrors = append(serverErrors, ginErr)
			}
		}

		if len(serverErrors) == 0 {
			c.Errors = nil
			if !c.Writer.Written() {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: last.Error()})
			}
			return
		}
		c.Errors = serverErrors

		// OWASP: no stack traces or driver messages in responses; the full
		// error stays in the logs and in APM.
		logger.FromContext(c.Request.Context()).Error("request failed", zap.Error(c.Errors.Last().Err))
		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
	}
}

// simulatedLatency is a route-scoped middleware that injects an artificial
// bottleneck recorded as its own APM span. Keeping it here means the payment
// handler has no knowledge of APM or of the demo delay.
func simulatedLatency(delay time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := sleepWithSpan(c.Request.Context(), delay); err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		c.Next()
	}
}

func sleepWithSpan(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	span, _ := apm.StartSpan(ctx, "simulated bottleneck", "app.internal.delay")
	defer span.End()

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
