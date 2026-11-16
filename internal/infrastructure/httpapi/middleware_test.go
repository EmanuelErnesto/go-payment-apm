package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
	"go.elastic.co/apm/v2/apmtest"
	"go.uber.org/zap/zapcore"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
	"github.com/EmanuelErnesto/go-payment-apm/internal/infrastructure/logger"
)

func newTestEngine(t *testing.T, handlerErr error) (*gin.Engine, *apmtest.RecordingTracer, *bytes.Buffer) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	tracer := apmtest.NewRecordingTracer()
	t.Cleanup(tracer.Close)

	var logs bytes.Buffer
	log := logger.New(logger.Options{Level: zapcore.InfoLevel, ServiceName: "test", Output: &logs})

	engine := gin.New()
	engine.Use(apmgin.Middleware(engine, apmgin.WithTracer(tracer.Tracer)), requestContext(log), requestLogger())
	engine.GET("/x", errorHandler(), func(c *gin.Context) { _ = c.Error(handlerErr) })
	return engine, tracer, &logs
}

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantStatus    int
		wantAPMErrors int
	}{
		{"domain error is a 400 and not an APM error", fmt.Errorf("bad amount: %w", domain.ErrInvalidInput), http.StatusBadRequest, 0},
		{"unexpected error is a 500 reported to APM", errors.New("db down"), http.StatusInternalServerError, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, tracer, logs := newTestEngine(t, tt.err)

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
			tracer.Flush(nil)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := len(tracer.Payloads().Errors); got != tt.wantAPMErrors {
				t.Fatalf("apm errors = %d, want %d", got, tt.wantAPMErrors)
			}

			// Every log line of the request must carry the APM correlation ids.
			for _, line := range bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n")) {
				var entry map[string]any
				if err := json.Unmarshal(line, &entry); err != nil {
					t.Fatalf("invalid json log line %q: %v", line, err)
				}
				if entry["trace.id"] == nil || entry["transaction.id"] == nil {
					t.Fatalf("log line without trace correlation: %s", line)
				}
			}
		})
	}
}
