package tui

import (
	"bufio"
	"fmt"
	"github.com/hunchulchoi/ollama-monitor-cli/internal/ollama"
	"os"
	"path/filepath"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TickMsg time.Time
type LogMsg struct {
	fileName string
	content  string
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type RunningModelInfo struct {
	Name          string
	Size          string
	VRAM          string
	ContextLength string
	TTL           string
}

type Model struct {
	client        *ollama.Client
	RunningModels []RunningModelInfo
	Logs          []*ollama.LogEntry
	Requests      []*ollama.LogEntry
	Latencies     []float64
	Stats         *ollama.ProcessStats
	CPUHistory    []float64
	MemHistory    []float64
	width         int
	height        int
}

func NewModel(client *ollama.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		doTick(),
		m.tailLogFile("server.log"),
		m.tailLogFile("app.log"),
	)
}

func (m Model) tailLogFile(name string) tea.Cmd {
	return func() tea.Msg {
		var logPath string
		customLogDir := os.Getenv("OLLAMA_LOG_DIR")

		if customLogDir != "" {
			logPath = filepath.Join(customLogDir, name)
		} else if runtime.GOOS == "windows" {
			localAppData := os.Getenv("LOCALAPPDATA")
			logPath = filepath.Join(localAppData, "Ollama", name)
		} else {
			home, _ := os.UserHomeDir()
			logPath = filepath.Join(home, ".ollama", "logs", name)
		}

		file, err := os.Open(logPath)
		if err != nil {
			return LogMsg{fileName: name, content: "Error opening " + name + ": " + err.Error()}
		}
		
		// Move to end of file
		file.Seek(0, 2)
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return LogMsg{fileName: name, content: line}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case TickMsg:
		res, err := m.client.GetRunningModels()
		if err == nil {
			var models []RunningModelInfo
			for _, mod := range res.Models {
				vramPercent := "0%"
				if mod.Size > 0 {
					vramPercent = fmt.Sprintf("%.0f%%", (float64(mod.SizeVRAM)/float64(mod.Size))*100)
				}

				ttl := "N/A"
				if mod.ExpiresAt != "" {
					exp, err := time.Parse(time.RFC3339, mod.ExpiresAt)
					if err == nil {
						diff := time.Until(exp)
						if diff > 0 {
							ttl = fmt.Sprintf("[%dm]", int(diff.Minutes()))
						} else {
							ttl = "[Expired]"
						}
					}
				}

				models = append(models, RunningModelInfo{
					Name:          mod.Name,
					Size:          fmt.Sprintf("%.1fGB", float64(mod.Size)/(1024*1024*1024)),
					VRAM:          "VRAM: " + vramPercent,
					ContextLength: fmt.Sprintf("CTX: %dk", mod.ContextLength/1024),
					TTL:           ttl,
				})
			}
			m.RunningModels = models
		}

		// Fetch process stats
		stats, _ := ollama.GetProcessStats()
		m.Stats = stats
		if stats != nil {
			m.CPUHistory = append(m.CPUHistory, stats.CPU)
			if len(m.CPUHistory) > 50 {
				m.CPUHistory = m.CPUHistory[len(m.CPUHistory)-50:]
			}
			m.MemHistory = append(m.MemHistory, stats.Memory)
			if len(m.MemHistory) > 50 {
				m.MemHistory = m.MemHistory[len(m.MemHistory)-50:]
			}
		}

		return m, doTick()
	case LogMsg:
		entry := ollama.ParseLine(msg.content)
		if entry != nil {
			if entry.RequestID != "" {
				m.Requests = append(m.Requests, entry)
				if len(m.Requests) > 8 {
					m.Requests = m.Requests[len(m.Requests)-8:]
				}
				if entry.ResponseTime > 0 {
					m.Latencies = append(m.Latencies, float64(entry.ResponseTime.Milliseconds()))
				}
			} else {
				m.Logs = append(m.Logs, entry)
				if len(m.Logs) > 15 {
					m.Logs = m.Logs[len(m.Logs)-15:]
				}
			}
		}
		return m, m.tailLogFile(msg.fileName)
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	contentWidth := m.width - 4
	boxStyle := BorderStyle.Width(contentWidth)

	header := HeaderStyle.Render(" 🦙 OLLAMA MONITOR") + "  " + time.Now().Format("15:04:05")
	if m.Stats != nil {
		header += fmt.Sprintf(" | CPU: %.1f%% | MEM: %.1fGB", m.Stats.CPU, m.Stats.Memory/(1024*1024*1024))
	}
	header += "\n"
	
	// Models Section
	modelsView := HeaderStyle.Render(" 📦 RUNNING MODELS") + "\n"
	if len(m.RunningModels) == 0 {
		modelsView += "  - None\n"
	} else {
		for _, info := range m.RunningModels {
			modelsView += fmt.Sprintf("  %-20s %-8s %-12s %-12s %s\n",
				info.Name, info.Size, info.VRAM, info.ContextLength, info.TTL)
		}
	}
	
	// Performance Section
	performanceView := LatencyStyle.Render(" ⚡ PERFORMANCE (Latency Flow)") + "\n"
	sparkline := RenderSparkline(m.Latencies, contentWidth-4)
	if sparkline == "No data" {
		performanceView += "  No data yet...\n"
	} else {
		performanceView += "  " + sparkline + "\n"
	}

	// Resources Section
	resourcesView := LatencyStyle.Render(" 📊 RESOURCE USAGE (History)") + "\n"
	cpuSpark := RenderSparkline(m.CPUHistory, (contentWidth/2)-6)
	memSpark := RenderSparkline(m.MemHistory, (contentWidth/2)-6)
	
	resourcesView += fmt.Sprintf("  CPU: %-25s  MEM: %-25s\n", cpuSpark, memSpark)
	
	// Requests Section
	requestsView := HeaderStyle.Render(" 🔄 RECENT REQUESTS") + "\n"
	if len(m.Requests) == 0 {
		requestsView += "  No requests yet...\n"
	} else {
		for _, req := range m.Requests {
			idShort := req.RequestID
			if len(idShort) > 8 {
				idShort = ".." + idShort[len(idShort)-8:]
			}
			requestsView += fmt.Sprintf("  ID: %s | %s | %s | %s | %s\n", 
				idShort, req.Method, req.Path, req.Status, req.ResponseTime.String())
		}
	}
	
	// Logs Section
	logsView := HeaderStyle.Render(" 📜 SERVER LOGS") + "\n"
	if len(m.Logs) == 0 {
		logsView += "  No logs yet...\n"
	} else {
		for _, log := range m.Logs {
			level := log.Level
			style := InfoStyle
			if level == "WARN" {
				style = WarnStyle
			} else if level == "ERROR" {
				style = ErrorStyle
			}
			msg := log.Msg
			if len(msg) > contentWidth-25 {
				msg = msg[:contentWidth-28] + "..."
			}
			timeStr := log.Time.Format("15:04:05")
			logsView += "  " + TimeStyle.Render(timeStr) + " " + style.Render(level) + " | " + msg + "\n"
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, 
		header, 
		boxStyle.Render(modelsView), 
		boxStyle.Render(performanceView),
		boxStyle.Render(resourcesView),
		boxStyle.Render(requestsView),
		boxStyle.Render(logsView),
	)
}
