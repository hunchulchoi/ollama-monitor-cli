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
		// Define custom colors for network metrics
		upColor := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))   // Orange Accent
		downColor := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Sky Blue Accent
		grayColor := lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Muted Gray for totals

		upSpeedStr := FormatBytes(m.UploadSpeed) + "/s"
		downSpeedStr := FormatBytes(m.DownloadSpeed) + "/s"
		upTotalStr := FormatBytes(float64(m.TotalUpload))
		downTotalStr := FormatBytes(float64(m.TotalDownload))

		networkPart := fmt.Sprintf(" | 🛜  %s %s %s | %s %s %s",
			upColor.Render("▲"), upSpeedStr, grayColor.Render("("+upTotalStr+")"),
			downColor.Render("▼"), downSpeedStr, grayColor.Render("("+downTotalStr+")"),
		)
		header += networkPart
	}
	if m.ShutdownActive {
		minutes := int(m.ShutdownDuration.Minutes())
		seconds := int(m.ShutdownDuration.Seconds()) % 60
		header += fmt.Sprintf(" | " + ErrorStyle.Bold(true).Render("SHUTDOWN IN %02d:%02d"), minutes, seconds)
	}
	return header
}
