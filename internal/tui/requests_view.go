package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderRequests(boxStyle lipgloss.Style, contentWidth int, maxRequests int) string {
	requestsView := HeaderStyle.Render(" 🔄 RECENT REQUESTS") + "\n"
	if len(m.Requests) == 0 {
		requestsView += "  No requests yet..."
	} else {
		start := len(m.Requests) - maxRequests
		if start < 0 {
			start = 0
		}
		displayReqs := m.Requests[start:]
		for i, req := range displayReqs {
			idShort := req.RequestID
			if len(idShort) > 8 {
				idShort = ".." + idShort[len(idShort)-8:]
			}
			
			// Dynamic path truncation to prevent line wrapping on narrow screens
			path := req.Path
			maxPathLen := contentWidth - 65
			if maxPathLen < 10 {
				maxPathLen = 10
			}
			if len(path) > maxPathLen {
				path = path[:maxPathLen-3] + "..."
			}

			timeStr := req.Time.Format("15:04:05")
			requestsView += fmt.Sprintf("  %s | ID: %s | %s | %s | %s | %s",
				TimeStyle.Render(timeStr), idShort, req.Method, path, req.Status, req.ResponseTime.String())
			if i < len(displayReqs)-1 {
				requestsView += "\n"
			}
		}
	}
	return boxStyle.Render(requestsView)
}
