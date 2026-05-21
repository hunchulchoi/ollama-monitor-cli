package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderDebugMetrics(boxStyle lipgloss.Style, contentWidth int, isFullMode bool) string {
	debugMetricsView := HeaderStyle.Render(" 🎯 DEBUG METRICS (Tokens & Speed)") + "\n"
	sparkWidth := (contentWidth - 16) / 2
	if sparkWidth < 5 {
		sparkWidth = 5
	}
	if isFullMode {
		tokenSpark := RenderSparkline(m.EvalTokens, sparkWidth, 10.0)
		tokenView := "  TOKENS: " + tokenSpark
		m.TPSChart.Draw()
		tpsView := "   TPS:\n" + m.TPSChart.View()
		debugMetricsView += lipgloss.JoinHorizontal(lipgloss.Top, tokenView, tpsView)
	} else {
		tokenSpark := RenderSparkline(m.EvalTokens, sparkWidth, 10.0)
		tpsSpark := RenderSparkline(m.TPSHistory, sparkWidth, 5.0)
		debugMetricsView += fmt.Sprintf("  TOKENS: %-*s  TPS: %-*s", sparkWidth, tokenSpark, sparkWidth, tpsSpark)
	}
	return boxStyle.Render(debugMetricsView)
}
