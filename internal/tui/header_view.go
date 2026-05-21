package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderHeader() string {
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
	return header
}
