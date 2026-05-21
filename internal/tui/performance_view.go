package tui

import (
	"fmt"
	"strings"

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
				latestStats = fmt.Sprintf(" [Latest Prompt: %d | Response: %d | Latency: %s]",
					lastReq.PromptEvalCount, lastReq.EvalCount, totalDuration.String())
			} else {
				latestStats = fmt.Sprintf(" [Latest Latency: %s]", totalDuration.String())
			}
		}
	}

	title := LatencyStyle.Render(" ⚡ NETWORK & PERFORMANCE")
	if latestStats != "" {
		title += lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(latestStats)
	}

	performanceView := title + "\n"
	if isFullMode {
		// Draw all three charts
		m.LatencyChart.Draw()
		m.UploadChart.Draw()
		m.DownloadChart.Draw()

		thirdWidth := (contentWidth - 6) / 3
		if thirdWidth < 10 {
			thirdWidth = 10
		}

		// Left Column: Latency Chart
		latTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render(fmt.Sprintf(" 🕒 API Latency Flow (Floor: 1s)"))
		latCol := latTitle + "\n" + m.LatencyChart.View()

		// Middle Column: Upload Speed Chart
		upSpeedStr := FormatBytes(m.UploadSpeed) + "/s"
		upTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true).Render(fmt.Sprintf(" 📤 Upload Speed: %s", upSpeedStr))
		upCol := upTitle + "\n" + m.UploadChart.View()

		// Right Column: Download Speed Chart
		downSpeedStr := FormatBytes(m.DownloadSpeed) + "/s"
		downTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Render(fmt.Sprintf(" 📥 Download Speed: %s", downSpeedStr))
		downCol := downTitle + "\n" + m.DownloadChart.View()

		// Combine horizontally
		hRow := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(thirdWidth).Render(latCol),
			lipgloss.NewStyle().Width(thirdWidth).Render("  "+upCol),
			lipgloss.NewStyle().Width(thirdWidth).Render("  "+downCol),
		)
		performanceView += hRow
	} else {
		// Compact mode: Sparklines
		sparkWidth := contentWidth - 40
		if sparkWidth < 10 {
			sparkWidth = 10
		}

		// Latency Sparkline
		latSpark := RenderSparkline(m.Latencies, sparkWidth, 1000.0)
		if latSpark == "No data" {
			latSpark = strings.Repeat(" ", sparkWidth)
		}
		performanceView += fmt.Sprintf("  Latency:  [ %s ] (Floor: 1s)\n", latSpark)

		// Upload Sparkline
		upSpark := RenderSparkline(m.UploadHistory, sparkWidth, 1.0)
		if upSpark == "No data" {
			upSpark = strings.Repeat(" ", sparkWidth)
		}
		performanceView += fmt.Sprintf("  Upload:   ↑ [ %s ] (Speed: %s/s | Total: %s)\n", 
			upSpark, FormatBytes(m.UploadSpeed), FormatBytes(float64(m.TotalUpload)))

		// Download Sparkline
		downSpark := RenderSparkline(m.DownloadHistory, sparkWidth, 1.0)
		if downSpark == "No data" {
			downSpark = strings.Repeat(" ", sparkWidth)
		}
		performanceView += fmt.Sprintf("  Download: ↓ [ %s ] (Speed: %s/s | Total: %s)", 
			downSpark, FormatBytes(m.DownloadSpeed), FormatBytes(float64(m.TotalDownload)))
	}

	return boxStyle.Render(performanceView)
}
