package tui

import (
	"bufio"
	"ollama-monitor/internal/ollama"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TickMsg time.Time
type LogMsg string

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type Model struct {
	client        *ollama.Client
	RunningModels []string
	Logs          []*ollama.LogEntry
	Latencies     []float64
}

func NewModel(client *ollama.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(doTick(), m.tailLogs())
}

func (m Model) tailLogs() tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		logPath := filepath.Join(home, ".ollama", "logs", "server.log")
		file, err := os.Open(logPath)
		if err != nil {
			return LogMsg("Error opening log: " + err.Error())
		}
		
		// Move to end of file
		file.Seek(0, 2)
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return LogMsg(line)
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case TickMsg:
		res, err := m.client.GetRunningModels()
		if err == nil {
			var names []string
			for _, mod := range res.Models {
				names = append(names, mod.Name)
			}
			m.RunningModels = names
		}
		return m, doTick()
	case LogMsg:
		entry := ollama.ParseLine(string(msg))
		if entry != nil {
			m.Logs = append(m.Logs, entry)
			if len(m.Logs) > 10 {
				m.Logs = m.Logs[len(m.Logs)-10:]
			}
			if entry.ResponseTime > 0 {
				m.Latencies = append(m.Latencies, float64(entry.ResponseTime.Milliseconds()))
			}
		}
		return m, m.tailLogs()
	}
	return m, nil
}

func (m Model) View() string {
	header := HeaderStyle.Render("OLLAMA MONITOR") + "\n"
	
	modelsView := "Running Models:\n"
	if len(m.RunningModels) == 0 {
		modelsView += "- None\n"
	} else {
		for _, name := range m.RunningModels {
			modelsView += "- " + name + "\n"
		}
	}
	
	performanceView := LatencyStyle.Render("Latency Flow: ") + RenderSparkline(m.Latencies, 20) + "\n"
	
	logsView := "Recent Logs:\n"
	for _, log := range m.Logs {
		level := log.Level
		style := InfoStyle
		if level == "WARN" {
			style = WarnStyle
		} else if level == "ERROR" {
			style = ErrorStyle
		}
		logsView += style.Render(level) + " | " + log.Msg + " | " + log.ResponseTime.String() + "\n"
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, header, modelsView, performanceView, BorderStyle.Render(logsView))
}
