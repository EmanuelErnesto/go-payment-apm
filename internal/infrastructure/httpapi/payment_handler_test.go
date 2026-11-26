package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateBindErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tooLarge := `{"description":"` + strings.Repeat("a", maxRequestBodyBytes) + `"}`

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{"body over the limit is a 413", tooLarge, http.StatusRequestEntityTooLarge, "request body too large"},
		{"malformed json is a 400", `{`, http.StatusBadRequest, "invalid request body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Binding fails before any use case runs, so none is needed.
			handler := NewPaymentHandler(nil, nil)
			engine := gin.New()
			engine.POST("/payments", limitBody(), handler.Create)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			var resp ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("invalid json response %q: %v", rec.Body.String(), err)
			}
			if resp.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", resp.Error, tt.wantError)
			}
		})
	}
}
