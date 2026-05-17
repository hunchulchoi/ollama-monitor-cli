# 🦙 Ollama Monitoring CLI

A high-performance, real-time TUI (Terminal User Interface) dashboard for monitoring your local [Ollama](https://ollama.com/) server. Built with Go and Bubble Tea.

![Ollama Monitor](docs/assets/screenshot.png)

## ✨ Features

- **📦 Model Monitoring:** Real-time view of running models, including size, VRAM usage, context length, and TTL.
- **⚡ Performance Visualizer:** Live sparkline charts showing response time (latency) trends for API requests.
- **📊 Resource Usage:** Historical CPU and Memory usage graphs for the Ollama process.
- **📜 Live Log Streaming:** Real-time tailing of `server.log` and `app.log` with color-coded log levels (INFO, WARN, ERROR).
- **🔄 Request Tracker:** Detailed table of recent API requests with IDs, methods, paths, and statuses.
- **📱 Responsive Layout:** Automatically adjusts UI components based on your terminal window size.

## 🚀 Quick Start

### Prerequisites

- [Go](https://go.dev/doc/install) 1.21 or higher.
- [Ollama](https://ollama.com/) installed and running on your machine.

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/hunchulchoi/ollama-monitor-cli.git
   cd ollama-monitor-cli
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Process Monitoring:** [gopsutil](https://github.com/shirou/gopsutil)

## 📂 Project Structure

- `main.go`: Application entry point.
- `internal/ollama/`: Core logic for API interaction, log parsing, and process stats.
- `internal/tui/`: TUI components, styles, and layout management.
- `docs/superpowers/`: Detailed design specifications and implementation plans.

## ⌨️ Keybindings

- `q` or `Ctrl+C`: Quit the application.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
Built with ❤️ for the Ollama community.
