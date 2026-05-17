package ollama

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	t.Run("Request Log", func(t *testing.T) {
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
		if entry.Method != "POST" {
			t.Errorf("Expected Method POST, got %s", entry.Method)
		}
		if entry.Path != "/api/v1/chat/new" {
			t.Errorf("Expected Path /api/v1/chat/new, got %s", entry.Path)
		}
		if entry.Status != "200" {
			t.Errorf("Expected Status 200, got %s", entry.Status)
		}
		if entry.ResponseTime.Seconds() < 30 {
			t.Errorf("Expected ResponseTime > 30s, got %v", entry.ResponseTime)
		}
	})

	t.Run("General Log", func(t *testing.T) {
		line := `time=2026-05-17T23:13:25.010+09:00 level=INFO msg="llama runner started in 15.15 seconds"`
		entry := ParseLine(line)
		if entry == nil {
			t.Fatal("Failed to parse log line")
		}
		if entry.Level != "INFO" {
			t.Errorf("Expected Level INFO, got %s", entry.Level)
		}
		if entry.Msg != "\"llama runner started in 15.15 seconds\"" {
			t.Errorf("Expected msg \"llama runner started in 15.15 seconds\", got %s", entry.Msg)
		}
		if entry.RequestID != "" {
			t.Errorf("Expected empty RequestID, got %s", entry.RequestID)
		}
	})
}
