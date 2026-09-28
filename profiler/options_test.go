package profiler

import (
	"testing"

	atatus "go.atatus.com/agent"
	"go.atatus.com/agent/profiler-internal/log"
)

func TestWithAPIKey(t *testing.T) {
	apiKey := "some-api-key"
	cfg := &config{}

	WithAPIKey(apiKey)(cfg)

	if cfg.apiKey != apiKey {
		t.Errorf("expected apiKey to be %q, got %q", apiKey, cfg.apiKey)
	}
}

func TestLicenseKeyFromDefaultTracer(t *testing.T) {
	prev := atatus.DefaultTracer.Service.LicenseKey
	atatus.DefaultTracer.Service.LicenseKey = "test-tracer-license-key"
	defer func() { atatus.DefaultTracer.Service.LicenseKey = prev }()

	cfg, err := defaultConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.apiKey != "test-tracer-license-key" {
		t.Errorf("expected apiKey to be %q, got %q", "test-tracer-license-key", cfg.apiKey)
	}
}

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected int
	}{
		{"debug", "debug", 0}, // log.LevelDebug = 0
		{"info", "info", 1},   // log.LevelInfo = 1
		{"warn", "warn", 2},   // log.LevelWarn = 2
		{"error", "error", 3}, // log.LevelError = 3
		{"upper", "DEBUG", 0},
		{"none", "none", 1},
		{"disable", "disable", 1},
		{"unknown", "trace", 1}, // Falls back to LevelInfo
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(log.LevelInfo) // Reset base log scale inside iteration since GetLevel queries this static scalar
			cfg := &config{}
			WithLogLevel(tt.level)(cfg)
			if int(log.GetLevel()) != tt.expected {
				t.Errorf("expected log level %d for %q, got %d", tt.expected, tt.level, log.GetLevel())
			}
		})
	}
}
