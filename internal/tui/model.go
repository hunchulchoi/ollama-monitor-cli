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
	linechart "github.com/NimbleMarkets/ntcharts/linechart/streamlinechart"
)

type TickMsg time.Time
type ShutdownTimerMsg time.Time
type LogMsg struct {
	fileName string
	content  string
	offset   int64
	err      error
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func doShutdownTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return ShutdownTimerMsg(t)
	})
}

type RunningModelInfo struct {
	Name          string
	Size          string
	VRAM          string
	ContextLength string
	TTL           string
}

type NewLogEntry *ollama.LogEntry

type Model struct {
	client           *ollama.Client
	ProxyChan        chan *ollama.LogEntry
	proxyServer      *ollama.ProxyServer
	RunningModels    []RunningModelInfo
	Logs             []*ollama.LogEntry
	Requests         []*ollama.LogEntry
	Latencies        []float64
	Stats            *ollama.ProcessStats
	CPUHistory       []float64
	MemHistory       []float64
	DebugMode        bool
	ProxyMode        bool
	EvalTokens       []float64 // Generated tokens per request
	TPSHistory       []float64 // Tokens Per Second
	width            int
	height           int
	ShutdownPending  bool
	ShutdownActive   bool
	ShutdownTime     time.Time
	ShutdownDuration time.Duration
	CPUChart         linechart.Model
	MemChart         linechart.Model
	LatencyChart     linechart.Model
	TPSChart         linechart.Model
}

func NewModel(client *ollama.Client, debugMode bool) *Model {
	cpuChart := linechart.New(20, 8)
	cpuChart.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Spring Green

	memChart := linechart.New(20, 8)
	memChart.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Deep Sky Blue

	latencyChart := linechart.New(20, 8)
	latencyChart.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange

	tpsChart := linechart.New(20, 8)
	tpsChart.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // Purple

	return &Model{
		client:       client,
		DebugMode:    debugMode,
		ProxyChan:    make(chan *ollama.LogEntry, 10),
		CPUChart:     cpuChart,
		MemChart:     memChart,
		LatencyChart: latencyChart,
		TPSChart:     tpsChart,
	}
}

func (m *Model) waitForProxyMetrics() tea.Cmd {
	return func() tea.Msg {
		entry := <-m.ProxyChan
		if entry == nil {
			return nil
		}
		return NewLogEntry(entry)
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		doTick(),
		m.tailLogFile("server.log", -1),
		m.tailLogFile("app.log", -1),
		m.waitForProxyMetrics(),
	)
}

func (m *Model) tailLogFile(name string, offset int64) tea.Cmd {
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
			time.Sleep(2 * time.Second)
			return LogMsg{fileName: name, content: "", offset: -1, err: err}
		}
		defer file.Close()
		
		info, err := file.Stat()
		if err != nil {
			time.Sleep(1 * time.Second)
			return LogMsg{fileName: name, content: "", offset: offset}
		}

		if offset == -1 || info.Size() < offset {
			offset, _ = file.Seek(0, 2)
		} else {
			file.Seek(offset, 0)
		}

		reader := bufio.NewReader(file)
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			return LogMsg{fileName: name, content: "", offset: offset}
		}

		return LogMsg{fileName: name, content: line, offset: offset + int64(len(line))}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShutdownPending {
			switch msg.String() {
			case "y", "Y":
				m.ShutdownPending = false
				m.ShutdownActive = true
				m.ShutdownDuration = 10 * time.Minute
				m.ShutdownTime = time.Now().Add(m.ShutdownDuration)
				return m, doShutdownTick()
			default:
				m.ShutdownPending = false
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "d", "D":
			m.DebugMode = !m.DebugMode
			m.EvalTokens = nil
			m.TPSHistory = nil
			// Restart Ollama in background
			go ollama.RestartOllama(m.DebugMode)
		case "p", "P":
			m.ProxyMode = !m.ProxyMode
			if m.ProxyMode {
				// Start proxy on 11435, pointing to 11434
				go func() {
					proxy, _ := ollama.NewProxyServer("http://localhost:11434", m.ProxyChan)
					m.proxyServer = proxy
					proxy.Start(":11435")
				}()
			} else {
				// Stop proxy
				if m.proxyServer != nil {
					m.proxyServer.Stop()
				}
			}
		case "s", "S":
			if !m.ShutdownActive {
				m.ShutdownPending = true
			} else {
				// Cancel shutdown
				m.ShutdownActive = false
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		chartWidth := (msg.Width - 10) / 2
		if chartWidth < 10 {
			chartWidth = 10
		}

		m.CPUChart.Resize(chartWidth, 8)
		m.MemChart.Resize(chartWidth, 8)

		latencyWidth := msg.Width - 6
		if latencyWidth < 10 {
			latencyWidth = 10
		}
		m.LatencyChart.Resize(latencyWidth, 8)
		m.TPSChart.Resize(chartWidth, 8)
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
						}
					}
				}

				models = append(models, RunningModelInfo{
					Name:          mod.Name,
					Size:          fmt.Sprintf("%.1fGB", float64(mod.Size)/(1024*1024*1024)),
					VRAM:          vramPercent,
					ContextLength: fmt.Sprintf("%d", mod.ContextLength),
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
			m.CPUChart.Push(stats.CPU)
			if len(m.CPUHistory) > 50 {
				m.CPUHistory = m.CPUHistory[len(m.CPUHistory)-50:]
			}
			m.MemHistory = append(m.MemHistory, stats.Memory)
			m.MemChart.Push(stats.Memory / (1024 * 1024 * 1024))
			if len(m.MemHistory) > 50 {
				m.MemHistory = m.MemHistory[len(m.MemHistory)-50:]
			}
		}

		return m, doTick()
	case ShutdownTimerMsg:
		if !m.ShutdownActive {
			return m, nil
		}
		m.ShutdownDuration = time.Until(m.ShutdownTime)
		if m.ShutdownDuration <= 0 {
			return m, tea.Quit
		}
		return m, doShutdownTick()
	case LogMsg:
		if msg.err != nil || msg.content == "" {
			return m, m.tailLogFile(msg.fileName, msg.offset)
		}

		entry := ollama.ParseLine(msg.content)
		if entry != nil {
			m.handleLogEntry(entry)
		}
		return m, m.tailLogFile(msg.fileName, msg.offset)
	case NewLogEntry:
		if msg != nil {
			m.handleLogEntry((*ollama.LogEntry)(msg))
		}
		return m, m.waitForProxyMetrics()
	}
	return m, nil
}

func (m *Model) handleLogEntry(entry *ollama.LogEntry) {
	if entry == nil {
		return
	}
	if entry.RequestID != "" {
		m.Requests = append(m.Requests, entry)
		if len(m.Requests) > 8 {
			m.Requests = m.Requests[len(m.Requests)-8:]
		}
		if entry.ResponseTime > 0 {
			m.Latencies = append(m.Latencies, float64(entry.ResponseTime.Milliseconds()))
			m.LatencyChart.Push(float64(entry.ResponseTime.Milliseconds()))
		}
	} else {
		m.Logs = append(m.Logs, entry)
		if len(m.Logs) > 15 {
			m.Logs = m.Logs[len(m.Logs)-15:]
		}

		// If it's a "finish generation" message with debug metrics
		if entry.EvalCount > 0 {
			m.EvalTokens = append(m.EvalTokens, float64(entry.EvalCount))
			if len(m.EvalTokens) > 50 {
				m.EvalTokens = m.EvalTokens[len(m.EvalTokens)-50:]
			}

			if entry.EvalDuration > 0 {
				tps := float64(entry.EvalCount) / entry.EvalDuration.Seconds()
				m.TPSHistory = append(m.TPSHistory, tps)
				m.TPSChart.Push(tps)
				if len(m.TPSHistory) > 50 {
					m.TPSHistory = m.TPSHistory[len(m.TPSHistory)-50:]
				}
			}
		}
	}
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	contentWidth := m.width - 4
	boxStyle := BorderStyle.Width(contentWidth)

	header := HeaderStyle.Render(" 🦙 OLLAMA MONITOR") + "  " + time.Now().Format("15:04:05")
	if m.Stats != nil {
		header += fmt.Sprintf(" | CPU: %.1f%% | MEM: %.1fGB", m.Stats.CPU, m.Stats.Memory/(1024*1024*1024))
	}
	if m.DebugMode {
		header += " | " + ErrorStyle.Bold(true).Render("DEBUG ON")
	}
	if m.ProxyMode {
		header += " | " + ErrorStyle.Bold(true).Foreground(lipgloss.Color("13")).Render("PROXY ON (Point Client to port 11435)")
	}
	if m.ShutdownActive {
		minutes := int(m.ShutdownDuration.Minutes())
		seconds := int(m.ShutdownDuration.Seconds()) % 60
		header += fmt.Sprintf(" | " + ErrorStyle.Bold(true).Render("SHUTDOWN IN %02d:%02d"), minutes, seconds)
	}

	// Models Section
	modelsView := HeaderStyle.Render(" 📦 RUNNING MODELS") + "\n"
	if len(m.RunningModels) == 0 {
		modelsView += "  - None"
	} else {
		for i, info := range m.RunningModels {
			modelsView += fmt.Sprintf("  %-20s %-8s %-12s %-12s %s",
				info.Name, info.Size, info.VRAM, info.ContextLength, info.TTL)
			if i < len(m.RunningModels)-1 {
				modelsView += "\n"
			}
		}
	}

	isFullMode := m.height >= 38

	// Debug Metrics Section (Conditional)
	var debugMetricsView string
	if m.DebugMode {
		debugMetricsView = HeaderStyle.Render(" 🎯 DEBUG METRICS (Tokens & Speed)") + "\n"
		if isFullMode {
			sparkWidth := (contentWidth - 16) / 2
			if sparkWidth < 5 {
				sparkWidth = 5
			}
			tokenSpark := RenderSparkline(m.EvalTokens, sparkWidth, 10.0)
			tokenView := "  TOKENS: " + tokenSpark
			tpsView := "   TPS:\n" + m.TPSChart.View()
			debugMetricsView += lipgloss.JoinHorizontal(lipgloss.Top, tokenView, tpsView)
		} else {
			sparkWidth := (contentWidth - 16) / 2
			if sparkWidth < 5 {
				sparkWidth = 5
			}
			tokenSpark := RenderSparkline(m.EvalTokens, sparkWidth, 10.0)
			tpsSpark := RenderSparkline(m.TPSHistory, sparkWidth, 5.0)
			debugMetricsView += fmt.Sprintf("  TOKENS: %-*s  TPS: %-*s", sparkWidth, tokenSpark, sparkWidth, tpsSpark)
		}
	}

	// Performance Section
	performanceView := LatencyStyle.Render(" ⚡ PERFORMANCE (Latency Flow)") + "\n"
	if isFullMode {
		performanceView += m.LatencyChart.View()
	} else {
		sparkline := RenderSparkline(m.Latencies, contentWidth-4, 1000.0) // Floor at 1s (1000ms)
		if sparkline == "No data" {
			performanceView += "  No data yet..."
		} else {
			performanceView += "  " + sparkline
		}
	}

	// Resources Section
	resourcesView := LatencyStyle.Render(" 📊 RESOURCE USAGE (History)") + "\n"
	if isFullMode {
		cpuView := "  CPU:\n" + m.CPUChart.View()
		memView := "   MEM:\n" + m.MemChart.View()
		resourcesView += lipgloss.JoinHorizontal(lipgloss.Top, cpuView, memView)
	} else {
		// Calculate dynamic width for sparklines to prevent overflow
		// Total available = contentWidth. We have labels "  CPU: " (7) and "  MEM: " (7)
		sparkWidth := (contentWidth - 16) / 2
		if sparkWidth < 5 {
			sparkWidth = 5
		}
		cpuSpark := RenderSparkline(m.CPUHistory, sparkWidth, 1.0)           // Floor at 1% CPU
		memSpark := RenderSparkline(m.MemHistory, sparkWidth, 1024*1024*1024) // Floor at 1GB RAM
		resourcesView += fmt.Sprintf("  CPU: %-*s  MEM: %-*s", sparkWidth, cpuSpark, sparkWidth, memSpark)
	}

	// Requests Section
	requestsView := HeaderStyle.Render(" 🔄 RECENT REQUESTS") + "\n"
	if len(m.Requests) == 0 {
		requestsView += "  No requests yet..."
	} else {
		for i, req := range m.Requests {
			idShort := req.RequestID
			if len(idShort) > 8 {
				idShort = ".." + idShort[len(idShort)-8:]
			}
			timeStr := req.Time.Format("15:04:05")
			requestsView += fmt.Sprintf("  %s | ID: %s | %s | %s | %s | %s",
				TimeStyle.Render(timeStr), idShort, req.Method, req.Path, req.Status, req.ResponseTime.String())
			if i < len(m.Requests)-1 {
				requestsView += "\n"
			}
		}
	}

	// Logs Section
	logsView := HeaderStyle.Render(" 📜 SERVER LOGS") + "\n"
	if len(m.Logs) == 0 {
		logsView += "  No logs yet..."
	} else {
		for i, log := range m.Logs {
			level := log.Level
			style := InfoStyle
			if level == "WARN" {
				style = WarnStyle
			} else if level == "ERROR" {
				style = ErrorStyle
			}
			msg := log.Msg
			if len(msg) > contentWidth-28 {
				msg = msg[:contentWidth-28] + "..."
			}
			timeStr := log.Time.Format("15:04:05")
			logsView += "  " + TimeStyle.Render(timeStr) + " " + style.Render(level) + " | " + msg
			if i < len(m.Logs)-1 {
				logsView += "\n"
			}
		}
	}

	footerText := " [q] Quit | [d] Toggle Debug | [p] Toggle Proxy | [s] Shutdown Timer"
	if m.ShutdownPending {
		footerText = ErrorStyle.Bold(true).Render(" 🚨 Shutdown in 10 mins? [y] Yes / [any] No ")
	} else if m.ShutdownActive {
		footerText = ErrorStyle.Bold(true).Render(" 🚨 Shutdown Timer Active (Press [s] to cancel) ")
	}
	footer := TimeStyle.Render(footerText)

	views := []string{
		header,
		boxStyle.Render(modelsView),
	}
	if m.DebugMode {
		views = append(views, boxStyle.Render(debugMetricsView))
	}
	views = append(views,
		boxStyle.Render(performanceView),
		boxStyle.Render(resourcesView),
		boxStyle.Render(requestsView),
		boxStyle.Render(logsView),
		footer,
	)

	return lipgloss.JoinVertical(lipgloss.Left, views...)
}
