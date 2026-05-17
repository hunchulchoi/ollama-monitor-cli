package main

import (
	"fmt"
	"ollama-monitor/internal/ollama"
	"ollama-monitor/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Prevent system sleep during monitoring
	ollama.PreventSleep()
	defer ollama.RestoreSleep()

	client := &ollama.Client{BaseURL: "http://localhost:11434"}
	p := tea.NewProgram(tui.NewModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
