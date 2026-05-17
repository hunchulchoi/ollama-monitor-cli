package main

import (
	"fmt"
	"ollama-monitor/internal/ollama"
	"ollama-monitor/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client := &ollama.Client{BaseURL: "http://localhost:11434"}
	p := tea.NewProgram(tui.NewModel(client))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
