package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderPerformance(boxStyle lipgloss.Style, contentWidth int, isFullMode bool) string {
	latestStats := ""
	if len(m.Requests) > 0 {
		lastReq := m.Requests[len(m.Requests)-1]
		totalDuration := lastReq.TotalDuration
		if totalDuration == 0 && lastReq.ResponseTime > 0 {
			totalDuration = lastReq.ResponseTime
		}
		if lastReq.EvalCount > 0 || lastReq.PromptEvalCount > 0 || totalDuration > 0 {
			if lastReq.PromptEvalCount > 0 || lastReq.EvalCount > 0 {
				latestStats = fmt.Sprintf(" [Latest: Prompt: %d | Response: %d | Duration: %s]",
					lastReq.PromptEvalCount, lastReq.EvalCount, totalDuration.String())
			} else {
				latestStats = fmt.Sprintf(" [Latest: Duration: %s]", totalDuration.String())
			}
		}
	}

	title := LatencyStyle.Render(" ⚡ PERFORMANCE (Latency Flow)")
	if latestStats != "" {
		title += lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(latestStats)
	}

	performanceView := title + "\n"
	if isFullMode {
		m.LatencyChart.Draw()
		performanceView += m.LatencyChart.View()
	} else {
		sparkline := RenderSparkline(m.Latencies, contentWidth-4, 1000.0) // Floor at 1s (1000ms)
		if sparkline == "No data" {
			performanceView += "  No data yet..."
		} else {
			performanceView += "  " + sparkline
		}
	}
	return boxStyle.Render(performanceView)
}
