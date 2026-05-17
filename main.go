package main

import (
	"fmt"
	"ollama-monitor/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.Model{})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
