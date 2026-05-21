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

func TestRenderDebugMetrics(t *testing.T) {
	model := NewModel(nil, true)
	model.EvalTokens = []float64{10, 20, 30}
	model.TPSHistory = []float64{15.5, 20.0}
	
	rendered := model.renderDebugMetrics(lipgloss.NewStyle(), 80, false)
	
	expectedTitle := "DEBUG METRICS"
	expectedTokens := "TOKENS:"
	expectedTPS := "TPS:"
	if !strings.Contains(rendered, expectedTitle) {
		t.Errorf("Expected debug metrics to contain '%s', got: %s", expectedTitle, rendered)
	}
	if !strings.Contains(rendered, expectedTokens) {
		t.Errorf("Expected debug metrics to contain '%s', got: %s", expectedTokens, rendered)
	}
	if !strings.Contains(rendered, expectedTPS) {
		t.Errorf("Expected debug metrics to contain '%s', got: %s", expectedTPS, rendered)
	}
}

func TestRenderResources(t *testing.T) {
	model := NewModel(nil, true)
	model.CPUHistory = []float64{2.5, 4.0}
	model.MemHistory = []float64{1024 * 1024 * 1024, 2 * 1024 * 1024 * 1024}
	
	rendered := model.renderResources(lipgloss.NewStyle(), 80, false)
	
	expectedTitle := "RESOURCE USAGE"
	expectedCPU := "CPU:"
	expectedMEM := "MEM:"
	if !strings.Contains(rendered, expectedTitle) {
		t.Errorf("Expected resources to contain '%s', got: %s", expectedTitle, rendered)
	}
	if !strings.Contains(rendered, expectedCPU) {
		t.Errorf("Expected resources to contain '%s', got: %s", expectedCPU, rendered)
	}
	if !strings.Contains(rendered, expectedMEM) {
		t.Errorf("Expected resources to contain '%s', got: %s", expectedMEM, rendered)
	}
}

func TestRenderRequests(t *testing.T) {
	model := NewModel(nil, true)
	model.Requests = []*ollama.LogEntry{
		{
			RequestID:    "xyz987",
			Time:         time.Now(),
			Method:       "GET",
			Path:         "/api/tags",
			Status:       "200",
			ResponseTime: 45 * time.Millisecond,
		},
	}
	rendered := model.renderRequests(lipgloss.NewStyle(), 80, 5)
	
	expectedTitle := "RECENT REQUESTS"
	expectedPath := "/api/tags"
	expectedStatus := "200"
	if !strings.Contains(rendered, expectedTitle) {
		t.Errorf("Expected requests view to contain '%s', got: %s", expectedTitle, rendered)
	}
	if !strings.Contains(rendered, expectedPath) {
		t.Errorf("Expected requests view to contain '%s', got: %s", expectedPath, rendered)
	}
	if !strings.Contains(rendered, expectedStatus) {
		t.Errorf("Expected requests view to contain '%s', got: %s", expectedStatus, rendered)
	}
}

func TestRenderLogs(t *testing.T) {
	model := NewModel(nil, true)
	model.Logs = []*ollama.LogEntry{
		{
			Time:  time.Now(),
			Level: "ERROR",
			Msg:   "Connection to database failed",
		},
	}
	rendered := model.renderLogs(lipgloss.NewStyle(), 80, 5)
	
	expectedTitle := "SERVER LOGS"
	expectedLevel := "ERROR"
	expectedMsg := "Connection to database failed"
	if !strings.Contains(rendered, expectedTitle) {
		t.Errorf("Expected logs view to contain '%s', got: %s", expectedTitle, rendered)
	}
	if !strings.Contains(rendered, expectedLevel) {
		t.Errorf("Expected logs view to contain '%s', got: %s", expectedLevel, rendered)
	}
	if !strings.Contains(rendered, expectedMsg) {
		t.Errorf("Expected logs view to contain '%s', got: %s", expectedMsg, rendered)
	}
}

func TestRenderFooter(t *testing.T) {
	model := NewModel(nil, true)
	model.ShutdownActive = true
	
	rendered := model.renderFooter()
	
	expectedText := "Shutdown Timer Active"
	if !strings.Contains(rendered, expectedText) {
		t.Errorf("Expected footer to contain '%s', got: %s", expectedText, rendered)
	}
}

func TestTUIBandwidthAccumulator(t *testing.T) {
	client := ollama.NewClient("http://localhost:11434")
	m := NewModel(client, false)

	// 1. Ingest dummy LogEntry with network sizes
	entry := &ollama.LogEntry{
		Level:        "METRIC",
		RequestSize:  1500,
		ResponseSize: 3000,
	}
	m.handleLogEntry(entry)

	if m.uploadTemp != 1500 || m.downloadTemp != 3000 {
		t.Errorf("Expected temp buffers 1500/3000, got %d/%d", m.uploadTemp, m.downloadTemp)
	}

	// 2. Perform tick update simulation
	_, _ = m.Update(BandwidthTickMsg(time.Now()))

	if m.UploadSpeed != 1500 || m.DownloadSpeed != 3000 {
		t.Errorf("Expected speeds 1500/3000, got %f/%f", m.UploadSpeed, m.DownloadSpeed)
	}
	if m.TotalUpload != 1500 || m.TotalDownload != 3000 {
		t.Errorf("Expected totals 1500/3000, got %d/%d", m.TotalUpload, m.TotalDownload)
	}
	if m.uploadTemp != 0 || m.downloadTemp != 0 {
		t.Errorf("Expected temp buffers to be reset, got %d/%d", m.uploadTemp, m.downloadTemp)
	}
}

func TestRenderHeaderNetworkMetrics(t *testing.T) {
	client := ollama.NewClient("http://localhost:11434")
	m := NewModel(client, false)
	m.ProxyMode = true
	m.UploadSpeed = 1024
	m.DownloadSpeed = 2048 * 1024
	m.TotalUpload = 4096
	m.TotalDownload = 8192 * 1024

	output := m.renderHeader()

	if !strings.Contains(output, "🛜") {
		t.Errorf("Expected header to contain network emoji 🛜, got: %s", output)
	}
	if !strings.Contains(output, "1.0 KB/s") {
		t.Errorf("Expected header to contain upload speed, got: %s", output)
	}
	if !strings.Contains(output, "2.0 MB/s") {
		t.Errorf("Expected header to contain download speed, got: %s", output)
	}
}

func TestNetworkChartsInitialization(t *testing.T) {
	m := NewModel(nil, true)
	if m.UploadChart.Style.GetForeground() == lipgloss.Color("") {
		t.Error("Expected UploadChart style to be initialized with a foreground color")
	}
	if m.DownloadChart.Style.GetForeground() == lipgloss.Color("") {
		t.Error("Expected DownloadChart style to be initialized with a foreground color")
	}
}

func TestNetworkChartsIngestionAndResize(t *testing.T) {
	m := NewModel(nil, true)
	m.uploadTemp = 1024 * 5   // 5 KB
	m.downloadTemp = 1024 * 10 // 10 KB

	// 1. Simulate bandwidth tick
	_, _ = m.Update(BandwidthTickMsg(time.Now()))
	
	if len(m.UploadHistory) != 1 || m.UploadHistory[0] != 5.0 {
		t.Errorf("Expected UploadHistory to have [5.0], got: %v", m.UploadHistory)
	}
	if len(m.DownloadHistory) != 1 || m.DownloadHistory[0] != 10.0 {
		t.Errorf("Expected DownloadHistory to have [10.0], got: %v", m.DownloadHistory)
	}

	// 2. Simulate resize event
	_, _ = m.Update(tea.WindowSizeMsg{Width: 96, Height: 40})
	
	// Layout calculations are correct and cause no panics.
}


