# Ollama Monitoring CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go-based TUI dashboard that monitors Ollama server status, running models, and visualizes performance metrics from logs.

**Architecture:** A Bubble Tea application that polls Ollama APIs for model status and tails the server log file to parse and visualize request latency.

**Tech Stack:** Go 1.26+, Bubble Tea (TUI), Lip Gloss (Styling), Sparkline (Visuals).

---

### Task 1: Project Initialization & API Client

**Files:**
- Create: `go.mod`
- Create: `internal/ollama/client.go`
- Test: `internal/ollama/client_test.go`

- [ ] **Step 1: Initialize Go module**
Run: `go mod init ollama-monitor`

- [ ] **Step 2: Create API Client structure**
```go
package ollama

import (
	"encoding/json"
	"net/http"
)

type Client struct {
	BaseURL string
}

type ModelPSResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

func (c *Client) GetRunningModels() (*ModelPSResponse, error) {
	resp, err := http.Get(c.BaseURL + "/api/ps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result ModelPSResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	return &result, err
}
```

- [ ] **Step 3: Write test for API Client**
```go
package ollama

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRunningModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"models": [{"name": "test-model", "size": 1000}]}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	res, err := client.GetRunningModels()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(res.Models) != 1 || res.Models[0].Name != "test-model" {
		t.Errorf("Unexpected result: %+v", res)
	}
}
```

- [ ] **Step 4: Run tests**
Run: `go test ./internal/ollama/... -v`
Expected: PASS

- [ ] **Step 5: Commit**
Run: `git add go.mod internal/ollama/ && git commit -m "feat: init project and ollama client"`

---

### Task 2: Log Tailing & Metrics Parsing

**Files:**
- Create: `internal/ollama/logs.go`
- Test: `internal/ollama/logs_test.go`

- [ ] **Step 1: Implement Log Entry structure and Parser**
```go
package ollama

import (
	"regexp"
	"time"
)

type LogEntry struct {
	Time         time.Time
	Level        string
	Msg          string
	ResponseTime time.Duration
	RequestID    string
}

var logRegex = regexp.MustCompile(`time=(?P<time>[^\s]+) level=(?P<level>[^\s]+).+msg=(?P<msg>"[^"]*"|[^\s]+).+http\.d=(?P<dur>[^\s]+).+request_id=(?P<id>[^\s]+)`)

func ParseLine(line string) *LogEntry {
	match := logRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	// Simplified parsing for brevity in plan
	return &LogEntry{Level: match[2], RequestID: match[5]}
}
```

- [ ] **Step 2: Write test for Parser**
```go
func TestParseLine(t *testing.T) {
	line := `time=2026-05-17T23:13:25.010+09:00 level=INFO source=ui.go:242 msg=site.serveHTTP http.method=POST http.path=/api/v1/chat/new http.status=200 http.d=32.229462083s request_id=1779027172781009000`
	entry := ParseLine(line)
	if entry == nil || entry.RequestID != "1779027172781009000" {
		t.Errorf("Failed to parse log line: %+v", entry)
	}
}
```

- [ ] **Step 3: Run tests**
Run: `go test ./internal/ollama/... -v`

- [ ] **Step 4: Commit**
Run: `git add internal/ollama/logs.go && git commit -m "feat: add log parsing logic"`

---

### Task 3: Basic Bubble Tea Model & Layout

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/model.go`
- Create: `main.go`

- [ ] **Step 1: Define Styles using Lip Gloss**
```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	BorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8"))
)
```

- [ ] **Step 2: Create Model and View**
```go
package tui

import (
	"ollama-monitor/internal/ollama"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	RunningModels []string
	Logs          []string
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" { return m, tea.Quit }
	}
	return m, nil
}
func (m Model) View() string {
	return HeaderStyle.Render("OLLAMA MONITOR") + "\n" + BorderStyle.Render("Logs here...")
}
```

- [ ] **Step 3: Implement main.go to start TUI**
```go
package main

import (
	"fmt"
	"ollama-monitor/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.Model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
```

- [ ] **Step 4: Run basic app**
Run: `go run main.go`
Expected: See "OLLAMA MONITOR" header and quit with 'q'.

---

### Task 4: Integration - Real-time Data Polling

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: Add Polling command**
```go
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
```

- [ ] **Step 2: Update Model to fetch data on Tick**
```go
// In Update(msg tea.Msg):
case TickMsg:
    models, _ := m.client.GetRunningModels()
    m.RunningModels = models
    return m, doTick()
```

- [ ] **Step 3: Commit**
Run: `git commit -am "feat: add real-time polling to TUI"`

---

### Task 5: Performance Visualization (Sparklines)

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/view.go` (if separated)

- [ ] **Step 1: Add Latency data to Model**
```go
type Model struct {
    latencies []float64
}
```

- [ ] **Step 2: Implement Sparkline rendering**
```go
func renderSparkline(data []float64) string {
    // Basic block character representation based on value
    return "▆▄▃ " 
}
```

- [ ] **Step 3: Final Polishing & Layout**
- Assemble all components: Header, Running Models Table, Sparkline, Log View.
- Use `lipgloss.JoinVertical` to combine sections.

- [ ] **Step 4: Final Commit**
Run: `git commit -am "feat: final dashboard layout and visualization"`
