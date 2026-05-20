package main

import (
	"flag"
	"fmt"
	"github.com/hunchulchoi/ollama-monitor-cli/internal/ollama"
	"github.com/hunchulchoi/ollama-monitor-cli/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Define command-line flags
	flagAPIURL := flag.String("url", "", "Ollama API URL (overrides .env OLLAMA_API_URL)")
	flagAPIKey := flag.String("key", "", "Ollama API Key (overrides .env OLLAMA_API_KEY)")
	flagLogDir := flag.String("logdir", "", "Ollama Log Directory (overrides .env OLLAMA_LOG_DIR)")
	flagDebug := flag.Bool("debug", false, "Start/Restart Ollama in Debug Mode (OLLAMA_DEBUG=1)")
	flag.Parse()

	debugMode := *flagDebug
	if debugMode {
		fmt.Println("\033[33m⚠️  Ollama will be restarted to apply Debug Mode settings.\033[0m")
		fmt.Println("\033[33m   (Any ongoing generation will be interrupted)\033[0m")
		fmt.Println("Restarting Ollama...")
		if err := ollama.RestartOllama(true); err != nil {
			fmt.Printf("Warning: Could not restart Ollama: %v\n", err)
		}
	} else {
		// Detect if Ollama is already running in debug mode
		debugMode = ollama.IsDebugMode()
	}

	// Priority: Flag > Environment Variable > Default
	apiURL := *flagAPIURL
	if apiURL == "" {
		apiURL = os.Getenv("OLLAMA_API_URL")
	}
	if apiURL == "" {
		apiURL = "http://localhost:11434"
	}

	apiKey := *flagAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("OLLAMA_API_KEY")
	}

	logDir := *flagLogDir
	if logDir == "" {
		logDir = os.Getenv("OLLAMA_LOG_DIR")
	}
	// If logDir is provided via flag or env, set it in the environment so tui package can pick it up
	if logDir != "" {
		os.Setenv("OLLAMA_LOG_DIR", logDir)
	}

	client := &ollama.Client{
		BaseURL: apiURL,
		APIKey:  apiKey,
	}

	p := tea.NewProgram(tui.NewModel(client, debugMode), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
