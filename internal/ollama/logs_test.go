package ollama

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	line := `time=2026-05-17T23:13:25.010+09:00 level=INFO source=ui.go:242 msg=site.serveHTTP http.method=POST http.path=/api/v1/chat/new http.status=200 http.d=32.229462083s request_id=1779027172781009000`
	entry := ParseLine(line)
	if entry == nil {
		t.Fatal("Failed to parse log line")
	}
	if entry.Level != "INFO" {
		t.Errorf("Expected Level INFO, got %s", entry.Level)
	}
	if entry.RequestID != "1779027172781009000" {
		t.Errorf("Expected ID 1779027172781009000, got %s", entry.RequestID)
	}
	if entry.ResponseTime.Seconds() < 30 {
		t.Errorf("Expected ResponseTime > 30s, got %v", entry.ResponseTime)
	}
}
