package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderLogs(boxStyle lipgloss.Style, contentWidth int, maxLogs int) string {
	logsView := HeaderStyle.Render(" 📜 SERVER LOGS") + "\n"
	if len(m.Logs) == 0 {
		logsView += "  No logs yet..."
	} else {
		start := len(m.Logs) - maxLogs
		if start < 0 {
			start = 0
		}
		displayLogs := m.Logs[start:]
		for i, log := range displayLogs {
			level := log.Level
			style := InfoStyle
			if level == "WARN" {
				style = WarnStyle
			} else if level == "ERROR" {
				style = ErrorStyle
			} else if level == "METRIC" {
				style = MetricStyle
			}
			msg := log.Msg
			maxLen := contentWidth - 28
			if maxLen < 5 {
				maxLen = 5
			}
			if len(msg) > maxLen && maxLen > 0 {
				msg = msg[:maxLen] + "..."
			}
			timeStr := log.Time.Format("15:04:05")
			logsView += "  " + TimeStyle.Render(timeStr) + " " + style.Render(level) + " | " + msg
			if i < len(displayLogs)-1 {
				logsView += "\n"
			}
		}
	}
	return boxStyle.Render(logsView)
}
