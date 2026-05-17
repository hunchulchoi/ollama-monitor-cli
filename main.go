package main

import (
	"fmt"
	"ollama-monitor/internal/ollama"
	"ollama-monitor/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	apiURL := os.Getenv("OLLAMA_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:11434"
	}

	apiKey := os.Getenv("OLLAMA_API_KEY")

	client := &ollama.Client{
		BaseURL: apiURL,
		APIKey:  apiKey,
	}
	p := tea.NewProgram(tui.NewModel(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
