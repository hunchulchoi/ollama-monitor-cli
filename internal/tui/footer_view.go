package tui

func (m *Model) renderFooter() string {
	footerText := " [q] Quit | [d] Toggle Debug | [p] Toggle Proxy | [s] Shutdown Timer | [r] Restart Ollama"
	if m.RestartPending {
		footerText = ErrorStyle.Bold(true).Render(" 🚨 Restart Ollama? [y] Yes / [any] No ")
	} else if m.ShutdownPending {
		footerText = ErrorStyle.Bold(true).Render(" 🚨 Shutdown in 10 mins? [y] Yes / [any] No ")
	} else if m.ShutdownActive {
		footerText = ErrorStyle.Bold(true).Render(" 🚨 Shutdown Timer Active (Press [s] to cancel) ")
	}
	return TimeStyle.Render(footerText)
}
