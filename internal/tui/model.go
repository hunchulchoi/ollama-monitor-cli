package tui

import (
	"ollama-monitor/internal/ollama"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type Model struct {
	client        *ollama.Client
	RunningModels []string
	Logs          []string
}

func NewModel(client *ollama.Client) Model {
	return Model{
		client: client,
	}
}

func (m Model) Init() tea.Cmd {
	return doTick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case TickMsg:
		res, err := m.client.GetRunningModels()
		if err == nil {
			var names []string
			for _, mod := range res.Models {
				names = append(names, mod.Name)
			}
			m.RunningModels = names
		}
		return m, doTick()
	}
	return m, nil
}

func (m Model) View() string {
	view := HeaderStyle.Render("OLLAMA MONITOR") + "\n"
	view += "Running Models:\n"
	for _, name := range m.RunningModels {
		view += "- " + name + "\n"
	}
	view += "\n" + BorderStyle.Render("Logs will appear here...")
	return view
}
