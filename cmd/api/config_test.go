package main

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestLoadConfig(t *testing.T) {
	t.Run("applies defaults and parses typed values", func(t *testing.T) {
		t.Setenv("DB_PASSWORD", "secret")
		t.Setenv("LOG_LEVEL", "DEBUG")
		t.Setenv("PAYMENT_SIMULATED_DELAY", "500ms")

		cfg, err := loadConfig()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.HTTP.Port != "8080" || cfg.HTTP.WriteTimeout != 30*time.Second || !cfg.HTTP.SwaggerEnabled {
			t.Fatalf("defaults not applied: %+v", cfg.HTTP)
		}
		if cfg.HTTP.SimulatedDelay != 500*time.Millisecond {
			t.Fatalf("delay = %v", cfg.HTTP.SimulatedDelay)
		}
		if cfg.Log.Level.Zap() != zapcore.DebugLevel {
			t.Fatalf("level = %v", cfg.Log.Level.Zap())
		}
		if cfg.APM.ServerURL.String() != "http://localhost:8200" || cfg.APM.SecretToken != "" {
			t.Fatalf("apm = %+v", cfg.APM)
		}
	})

	errorCases := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"missing required DB_PASSWORD", map[string]string{"DB_PASSWORD": ""}, "DB_PASSWORD"},
		{"invalid duration", map[string]string{"HTTP_WRITE_TIMEOUT": "soon"}, "HTTP_WRITE_TIMEOUT"},
		{"invalid log level", map[string]string{"LOG_LEVEL": "loud"}, "LOG_LEVEL"},
		{"relative APM URL", map[string]string{"ELASTIC_APM_SERVER_URL": "apm-server:8200"}, "ELASTIC_APM_SERVER_URL"},
		{"delay not lower than write timeout", map[string]string{"PAYMENT_SIMULATED_DELAY": "30s"}, "PAYMENT_SIMULATED_DELAY"},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DB_PASSWORD", "secret")
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			_, err := loadConfig()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want mention of %s", err, tc.wantErr)
			}
		})
	}
}
