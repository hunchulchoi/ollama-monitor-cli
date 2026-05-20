package ollama

import (
	"testing"
)

func TestGetProcessStats(t *testing.T) {
	stats, err := GetProcessStats()
	if err != nil {
		t.Fatalf("Unexpected error from GetProcessStats: %v", err)
	}

	if stats != nil {
		t.Logf("Found running Ollama process(es). CPU: %f, Memory: %f MB", stats.CPU, stats.Memory/(1024*1024))
		if stats.CPU < 0 {
			t.Errorf("Unexpected negative CPU value: %f", stats.CPU)
		}
		if stats.Memory < 0 {
			t.Errorf("Unexpected negative Memory value: %f", stats.Memory)
		}
	} else {
		t.Log("No running Ollama processes found on this system.")
	}
}
