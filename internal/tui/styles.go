package tui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	BorderStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("8"))
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	WarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	LatencyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
)
