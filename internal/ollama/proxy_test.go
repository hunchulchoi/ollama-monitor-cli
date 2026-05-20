package ollama

import (
	"encoding/json"
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
