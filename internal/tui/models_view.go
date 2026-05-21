package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderRunningModels(boxStyle lipgloss.Style, contentWidth int) string {
	modelsView := HeaderStyle.Render(" 📦 RUNNING MODELS") + "\n"
	if m.APIError != nil {
		modelsView += "  " + ErrorStyle.Bold(true).Render(fmt.Sprintf("⚠️  API ERROR: %v", m.APIError))
	} else if len(m.RunningModels) == 0 {
		modelsView += "  - None"
	} else {
		headerLine := fmt.Sprintf("  %-20s %-8s %-12s %-12s %s", "MODEL", "SIZE", "VRAM", "CTX LIMIT", "TTL")
		headerLine = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Bold(true).Render(headerLine)
		modelsView += headerLine + "\n"

		for i, info := range m.RunningModels {
			line := fmt.Sprintf("  %-20s %-8s %-12s %-12s %s",
				info.Name, info.Size, info.VRAM, info.ContextLength, info.TTL)
			maxLen := contentWidth - 4
			if maxLen < 10 {
				maxLen = 10
			}
			if len(line) > maxLen {
				line = line[:maxLen-3] + "..."
			}
			modelsView += line
			if i < len(m.RunningModels)-1 {
				modelsView += "\n"
			}
		}
	}
	return boxStyle.Render(modelsView)
}
