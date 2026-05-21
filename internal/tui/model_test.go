package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func TestRenderPerformance(t *testing.T) {
	model := NewModel(nil, true)
	model.Requests = []*ollama.LogEntry{
		{
			RequestID:       "1234567890",
			Time:            time.Now(),
			Method:          "POST",
			Path:            "/api/generate",
			Status:          "200",
			PromptEvalCount: 50,
			EvalCount:       150,
			TotalDuration:   3500 * time.Millisecond,
		},
	}

	boxStyle := lipgloss.NewStyle()
	rendered := model.renderPerformance(boxStyle, 80, true)

	expectedPrompt := "Prompt: 50"
	expectedResponse := "Response: 150"
	expectedDuration := "Duration: 3.5s"

	if !strings.Contains(rendered, expectedPrompt) {
		t.Errorf("Expected performance rendering to contain '%s', got: %s", expectedPrompt, rendered)
	}
	if !strings.Contains(rendered, expectedResponse) {
		t.Errorf("Expected performance rendering to contain '%s', got: %s", expectedResponse, rendered)
	}
	if !strings.Contains(rendered, expectedDuration) {
		t.Errorf("Expected performance rendering to contain '%s', got: %s", expectedDuration, rendered)
	}
}

func TestRestartOllamaConfirm(t *testing.T) {
	model := NewModel(nil, true)

	// 1. Verify key 'r' or 'R' sets RestartPending to true
	m, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	updatedModel := m.(*Model)
	if !updatedModel.RestartPending {
		t.Error("Expected RestartPending to be true after pressing 'r'")
	}

	// 2. Verify footer rendering when RestartPending is true
	footer := updatedModel.renderFooter()
	expectedConfirmMsg := "Restart Ollama? [y] Yes"
	if !strings.Contains(footer, expectedConfirmMsg) {
		t.Errorf("Expected footer to contain confirm message '%s', got: %s", expectedConfirmMsg, footer)
	}

	// 3. Verify key 'y' or 'Y' inside RestartPending clears RestartPending
	m2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	updatedModel2 := m2.(*Model)
	if updatedModel2.RestartPending {
		t.Error("Expected RestartPending to be cleared to false after pressing 'y'")
	}

	// 4. Verify any other key clears RestartPending
	updatedModel.RestartPending = true
	m3, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	updatedModel3 := m3.(*Model)
	if updatedModel3.RestartPending {
		t.Error("Expected RestartPending to be cleared to false after pressing 'n' (any other key)")
	}
}

func TestRenderHeader(t *testing.T) {
	model := NewModel(nil, true)
	model.Stats = &ollama.ProcessStats{
		CPU:    45.2,
		Memory: 2 * 1024 * 1024 * 1024,
	}
	rendered := model.renderHeader()
	
	expectedTitle := "🦙 OLLAMA MONITOR"
	expectedStats := "CPU: 45.2% | MEM: 2.0GB"
	if !strings.Contains(rendered, expectedTitle) {
		t.Errorf("Expected header to contain '%s', got: %s", expectedTitle, rendered)
	}
	if !strings.Contains(rendered, expectedStats) {
		t.Errorf("Expected header to contain '%s', got: %s", expectedStats, rendered)
	}
}

func TestRenderRunningModels(t *testing.T) {
	model := NewModel(nil, true)
	model.RunningModels = []RunningModelInfo{
		{
			Name:          "phi3:latest",
			Size:          "2.2GB",
			VRAM:          "100%",
			ContextLength: "4k",
			TTL:           "[8m]",
		},
	}
	rendered := model.renderRunningModels(lipgloss.NewStyle(), 80)
	
	expectedName := "phi3:latest"
	expectedSize := "2.2GB"
	if !strings.Contains(rendered, expectedName) {
		t.Errorf("Expected running models to contain '%s', got: %s", expectedName, rendered)
	}
	if !strings.Contains(rendered, expectedSize) {
		t.Errorf("Expected running models to contain '%s', got: %s", expectedSize, rendered)
	}
}


