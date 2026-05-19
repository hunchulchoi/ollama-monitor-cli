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

	t.Run("GIN Log", func(t *testing.T) {
		line := `[GIN] 2026/05/18 - 21:07:28 | 200 |    5.350416ms |       127.0.0.1 | GET      "/api/tags"`
		entry := ParseLine(line)
		if entry == nil {
			t.Fatal("Failed to parse GIN log line")
		}
		if entry.Status != "200" {
			t.Errorf("Expected Status 200, got %s", entry.Status)
		}
		if entry.Method != "GET" {
			t.Errorf("Expected Method GET, got %s", entry.Method)
		}
		if entry.Path != "/api/tags" {
			t.Errorf("Expected Path /api/tags, got %s", entry.Path)
		}
		if entry.ResponseTime.Milliseconds() < 5 {
			t.Errorf("Expected ResponseTime > 5ms, got %v", entry.ResponseTime)
		}
	})

	t.Run("Debug Generation Log", func(t *testing.T) {
		line := `time=2024-10-09T13:15:55.911+08:00 level=INFO source=sched.go:714 msg="finish generation" "prompt_eval_count": 32, "prompt_eval_duration": 1370000000, "eval_count": 279, "eval_duration": 5591000000, "total_duration": 7012000000, "load_duration": 1234567`
		entry := ParseLine(line)
		if entry == nil {
			t.Fatal("Failed to parse debug log line")
		}
		if entry.PromptEvalCount != 32 {
			t.Errorf("Expected prompt_eval_count 32, got %d", entry.PromptEvalCount)
		}
		if entry.EvalCount != 279 {
			t.Errorf("Expected eval_count 279, got %d", entry.EvalCount)
		}
		if entry.PromptEvalDuration.Seconds() != 1.37 {
			t.Errorf("Expected prompt_eval_duration 1.37s, got %v", entry.PromptEvalDuration)
		}
		if entry.EvalDuration.Seconds() != 5.591 {
			t.Errorf("Expected eval_duration 5.591s, got %v", entry.EvalDuration)
		}
		if entry.TotalDuration.Seconds() != 7.012 {
			t.Errorf("Expected total_duration 7.012s, got %v", entry.TotalDuration)
		}
		if entry.LoadDuration.Nanoseconds() != 1234567 {
			t.Errorf("Expected load_duration 1234567ns, got %v", entry.LoadDuration)
		}
	})
}
