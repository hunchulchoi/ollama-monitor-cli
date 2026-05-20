package tui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
	BorderStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	WarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	MetricStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	LatencyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Padding(0, 1)
	TimeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
)
