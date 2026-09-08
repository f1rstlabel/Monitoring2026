package monitoring

import (
	"context"
	"testing"

	"sanoc/backend/internal/domain"
)

func TestValidateTargetURL(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{name: "public https", target: "https://example.com/health", wantErr: false},
		{name: "public http with port", target: "http://example.com:8080/health", wantErr: false},
		{name: "missing scheme", target: "example.com/health", wantErr: true},
		{name: "unsupported scheme", target: "ftp://example.com/health", wantErr: true},
		{name: "localhost", target: "http://localhost:8080/health", wantErr: true},
		{name: "private IP", target: "http://192.168.1.10/health", wantErr: true},
		{name: "embedded credentials", target: "https://user:password@example.com/health", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTargetURL(tt.target); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTargetURL(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestPublicMonitorFailureProducesDownResult(t *testing.T) {
	result := RunPublicMonitorCheck(context.Background(), domain.PublicMonitor{
		MonitorType:    domain.PublicMonitorHTTP,
		TargetURL:      "https://sanoc-public-monitor-test.invalid/health",
		TimeoutSeconds: 1,
		RetryCount:     0,
	})
	if result.Status != domain.PublicMonitorDown {
		t.Fatalf("failure result status = %q, want DOWN", result.Status)
	}
	if result.Error == "" {
		t.Fatal("failure result should include an actionable error")
	}
}

func TestLookupJSONPath(t *testing.T) {
	payload := map[string]interface{}{"data": map[string]interface{}{"status": "ok"}}
	value, ok := lookupJSONPath(payload, "data.status")
	if !ok || value != "ok" {
		t.Fatalf("lookupJSONPath() = %v, %v; want ok, true", value, ok)
	}
}
