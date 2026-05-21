package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderResources(boxStyle lipgloss.Style, contentWidth int, isFullMode bool) string {
	resourcesView := LatencyStyle.Render(" 📊 RESOURCE USAGE (History)") + "\n"
	if isFullMode {
		m.CPUChart.Draw()
		m.MemChart.Draw()
		cpuView := "  CPU:\n" + m.CPUChart.View()
		memView := "   MEM:\n" + m.MemChart.View()
		resourcesView += lipgloss.JoinHorizontal(lipgloss.Top, cpuView, memView)
	} else {
		sparkWidth := (contentWidth - 16) / 2
		if sparkWidth < 5 {
			sparkWidth = 5
		}
		cpuSpark := RenderSparkline(m.CPUHistory, sparkWidth, 1.0)           // Floor at 1% CPU
		memSpark := RenderSparkline(m.MemHistory, sparkWidth, 1024*1024*1024) // Floor at 1GB RAM
		resourcesView += fmt.Sprintf("  CPU: %-*s  MEM: %-*s", sparkWidth, cpuSpark, sparkWidth, memSpark)
	}
	return boxStyle.Render(resourcesView)
}
