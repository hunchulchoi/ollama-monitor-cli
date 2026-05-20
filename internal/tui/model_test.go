package tui

import (
	"fmt"
	"testing"
	"time"

	"github.com/hunchulchoi/ollama-monitor-cli/internal/ollama"
)

func TestModelViewPanics(t *testing.T) {
	model := NewModel(nil, true)
	
	model.Requests = []*ollama.LogEntry{
		{
			RequestID:    "1234567890",
			Time:         time.Now(),
			Method:       "POST",
			Path:         "/api/generate",
			Status:       "200",
			ResponseTime: 500 * time.Millisecond,
		},
	}
	model.Logs = []*ollama.LogEntry{
		{
			Time:  time.Now(),
			Level: "INFO",
			Msg:   "Some server log message that is quite long to trigger truncation logic.",
		},
	}
	model.RunningModels = []RunningModelInfo{
		{
			Name:          "llama3:8b",
			Size:          "4.7GB",
			VRAM:          "100%",
			ContextLength: "8192",
			TTL:           "[9m]",
		},
	}
	model.Stats = &ollama.ProcessStats{
		CPU:    12.5,
		Memory: 4 * 1024 * 1024 * 1024,
	}

	dimensions := []struct{ w, h int }{
		{80, 40},  // Standard full
		{80, 20},  // Standard compact
		{40, 15},  // Narrow compact
		{15, 8},   // Very narrow
		{5, 2},    // Extremely tiny
		{0, 0},    // Zero/uninitialized
	}
	
	for _, dim := range dimensions {
		t.Run(fmt.Sprintf("%dx%d", dim.w, dim.h), func(t *testing.T) {
			model.width = dim.w
			model.height = dim.h
			
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("View() panicked for size %dx%d: %v", dim.w, dim.h, r)
				}
			}()
			
			_ = model.View()
		})
	}
}
