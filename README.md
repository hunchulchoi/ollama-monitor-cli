# 🦙 Ollama Monitoring CLI

A high-performance, real-time TUI (Terminal User Interface) dashboard for monitoring your local [Ollama](https://ollama.com/) server. Built with Go and Bubble Tea. Supports macOS, Linux, and Windows.

![Ollama Monitor](docs/assets/screenshot.png)

## ✨ Features

- **📦 Model Monitoring:** Real-time view of running models, featuring a structured table header (**MODEL**, **SIZE**, **VRAM**, **CTX LIMIT**, **TTL**).
  - Context limits are beautifully formatted into human-readable Kilo-notation (e.g., `8k`, `256k`).
- **⚡ Performance Visualizer:** Live latency flow showing response time trends. Includes real-time token telemetry showing prompt and response token counts:
  - `[Latest: Prompt: X | Response: Y | Duration: Z]`
- **🎯 Debug Metrics:** Visualizes token generation speed (TPS - Tokens Per Second) using stream-line charts and response token sparklines. (Activated via `d` key).
- **🔌 Built-in Proxy Server:** Capture precise API telemetry by routing your OpenAI/Ollama clients through the built-in proxy on port `11435`.
  - Seamlessly handles proxy network errors and routes them directly into the TUI server logs without cluttering the screen.
- **📊 Resource Usage:** Historical CPU and Memory usage graphs for the Ollama process, automatically summing up master daemon (`ollama serve`) and inference engines (`ollama runner`).
- **📜 Live Log Streaming:** Real-time tailing of `server.log` and `app.log` with color-coded log levels (INFO, WARN, ERROR, and captured proxy **METRIC** logs) and timestamps.
- **🔄 Request Tracker:** Detailed table of recent API requests with IDs, methods, paths, and statuses.
- **📱 Responsive Layout:** Automatically adjusts UI components and chart bounds dynamically based on your terminal window size.

## 🚀 Quick Start

### Prerequisites

- [Go](https://go.dev/doc/install) 1.21 or higher.
- [Ollama](https://ollama.com/) installed and running.

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

## 🔌 Proxy Setup (Highly Recommended)

To get precise token metrics (Prompt/Response counts) and real-time TPS without restarting Ollama, you can use the built-in Proxy:

1. Press **`p`** in the TUI to turn on the proxy. The header will show: `PROXY ON (Point Client to port 11435)`.
2. Configure your API client (e.g., Open WebUI, Python SDK, or curl) to point to port `11435` instead of `11434`.
   * **Python SDK Example:**
     ```python
     from ollama import Client
     client = Client(host='http://localhost:11435')
     response = client.chat(model='qwen2.5:32b', messages=[...])
     ```
3. Watch the precise `Prompt`, `Response` token count, and `TPS` appear instantly in the TUI!

## ⚙️ Configuration

The application can be configured using command-line flags or a `.env` file.

### Priority
1. Command-line Flags (Highest)
2. `.env` File
3. System Defaults (Lowest)

### Command-line Flags
```bash
Usage of ollama-monitor:
  -url string
        Ollama API URL (e.g., http://localhost:11434)
  -key string
        Ollama API Key (if using a proxy/auth)
  -logdir string
        Custom Ollama Log Directory path
```

### Environment Variables (.env)
Copy the example file and modify it:
```bash
cp .env.example .env
```
Available variables:
- `OLLAMA_API_URL`: Ollama API endpoint.
- `OLLAMA_API_KEY`: API authentication key.
- `OLLAMA_LOG_DIR`: Custom path to the directory containing `server.log` and `app.log`.

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Charts:** [ntcharts](https://github.com/NimbleMarkets/ntcharts)
- **Process Monitoring:** [gopsutil](https://github.com/shirou/gopsutil)

## ⌨️ Keybindings

- `q` or `Ctrl+C`: **Quit** the application.
- `d`: **Toggle Debug Mode**. Re-runs Ollama with debug output, enabling the `🎯 DEBUG METRICS` charts.
- `p`: **Toggle Proxy Mode**. Starts/stops the built-in telemetry proxy server on port `11435`.
- `r`: **Restart Ollama**. Safely restarts the Ollama process with a confirmation dialog (`🚨 Restart Ollama? [y]/[any]`).
- `s`: **Shutdown Timer**. Schedules a safe system shutdown in 10 minutes (useful for overnight/long-running runs) or cancels an active shutdown timer.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
Built with ❤️ for the Ollama community.
