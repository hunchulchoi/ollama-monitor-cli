package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	RunningModels []string
	Logs          []string
}

func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m Model) View() string {
	return HeaderStyle.Render("OLLAMA MONITOR") + "\n" + BorderStyle.Render("Logs will appear here...")
}
