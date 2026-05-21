package ollama

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxyMetricsUnmarshal(t *testing.T) {
	jsonData := `{
		"model": "llama3",
		"created_at": "2026-05-20T14:57:38Z",
		"response": "Hello!",
		"done": true,
		"context": [1, 2, 3],
		"total_duration": 7012000000,
		"load_duration": 1234567,
		"prompt_eval_count": 32,
		"prompt_eval_duration": 1370000000,
		"eval_count": 279,
		"eval_duration": 5591000000
	}`

	var metrics ProxyMetrics
	err := json.Unmarshal([]byte(jsonData), &metrics)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if metrics.Model != "llama3" {
		t.Errorf("Expected model 'llama3', got '%s'", metrics.Model)
	}
	if metrics.EvalCount != 279 {
		t.Errorf("Expected eval_count 279, got %d", metrics.EvalCount)
	}
	if metrics.PromptEvalCount != 32 {
		t.Errorf("Expected prompt_eval_count 32, got %d", metrics.PromptEvalCount)
	}
	if metrics.EvalDuration != 5591*time.Millisecond {
		t.Errorf("Expected eval_duration 5.591s, got %v", metrics.EvalDuration)
	}
	if metrics.PromptEvalDuration != 1370*time.Millisecond {
		t.Errorf("Expected prompt_eval_duration 1.37s, got %v", metrics.PromptEvalDuration)
	}
	if metrics.TotalDuration != 7012*time.Millisecond {
		t.Errorf("Expected total_duration 7.012s, got %v", metrics.TotalDuration)
	}
}

func TestProxyByteMeasurement(t *testing.T) {
	// 1. Create dummy target server
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response-body-12345")) // 19 bytes response body
	}))
	defer target.Close()

	metricsOut := make(chan *LogEntry, 10)
	proxy, err := NewProxyServer(target.URL, metricsOut)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Mock request
	req := httptest.NewRequest("POST", "/api/tags", bytes.NewBuffer([]byte("req-body-123"))) // 12 bytes request body
	req.Header.Set("X-Request-ID", "test-network-id")
	rec := httptest.NewRecorder()

	// 3. Directly call ServeHTTP to trigger traffic measurement handler
	proxy.ServeHTTP(rec, req)

	// 4. Verify metric sent to channel
	select {
	case entry := <-metricsOut:
		if entry.RequestID != "test-network-id" {
			t.Errorf("Expected RequestID 'test-network-id', got '%s'", entry.RequestID)
		}
		if entry.RequestSize <= 12 {
			t.Errorf("Expected RequestSize > 12 (body + header), got %d", entry.RequestSize)
		}
		if entry.ResponseSize <= 19 {
			t.Errorf("Expected ResponseSize > 19 (body + header), got %d", entry.ResponseSize)
		}
		if entry.Level != "METRIC" {
			t.Errorf("Expected Level 'METRIC', got '%s'", entry.Level)
		}
	default:
		t.Fatal("Expected LogEntry in metricsOut channel, but timed out")
	}
}

