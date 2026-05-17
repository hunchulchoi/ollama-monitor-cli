# Design Spec: Ollama Monitoring CLI (Go + Bubble Tea)

## 1. Overview
A high-performance TUI (Terminal User Interface) dashboard for monitoring Ollama local LLM server. It provides real-time insights into loaded models, resource usage, and visualized performance metrics derived from system logs.

## 2. Goals
- Real-time monitoring of loaded models and system resources.
- Visual representation of response times and request history.
- Live streaming and parsing of Ollama server/app logs.
- Single binary distribution using Go.

## 3. Architecture
- **Language:** Go (Golang)
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture for TUI)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Data Source:**
  - Ollama REST API (`/api/ps`, `/api/tags`)
  - Local Log Files (`~/.ollama/logs/server.log`, `app.log`)

## 4. UI Components

### 4.1 Header
- Title: "OLLAMA MONITOR"
- Server Status: Online/Offline (color-coded)
- System Info: Current Time, CPU/Memory Usage (optional)

### 4.2 Model Monitor (Top)
- Displays currently loaded models in VRAM.
- Columns: Name, Size, VRAM Usage, Context Length, TTL.

### 4.3 Performance Visualizer (Middle)
- **Latency Sparkline:** A visual chart showing response times (`http.d`) from the last 20 requests.
- **Request ID Tracker:** A summarized table of the most recent API requests (ID, Method, Path, Status, Latency).

### 4.4 Log Viewer (Bottom)
- Real-time tailing of `server.log`.
- Color-coded levels: INFO (Blue), WARN (Yellow), ERROR (Red).
- Auto-scrolling to the latest entry.

## 5. Technical Requirements

### 5.1 Log Parsing Logic
- Regex-based parsing for structured logs:
  - Time: `time=...`
  - Level: `level=...`
  - Response Time: `http.d=...`
  - Request ID: `request_id=...`
- Extraction of duration strings (e.g., `32.2s`, `1.6ms`) converted to milliseconds for graphing.

### 5.2 Polling & Updates
- Ollama API polling: Every 2 seconds.
- Log tailing: Non-blocking file watcher (or polling `io.Seek`).

## 6. Layout Mockup
```text
┌────────────────────────────────────────────────────────────┐
│ OLLAMA MONITOR [ONLINE]                           23:14:00 │
├────────────────────────────────────────────────────────────┤
│ RUNNING MODELS                                             │
│ gemma4:26b      25.8GB    VRAM: 100%    CTX: 262k   [5m]   │
├────────────────────────────────────────────────────────────┤
│ PERFORMANCE (Latency Flow)                                 │
│ 32.2s [▆       ] 1.6ms [_       ]                          │
├────────────────────────────────────────────────────────────┤
│ RECENT REQUESTS                                            │
│ ID: ..846e79 | GET  | /api/v1/chat/id | 200 | 1.6ms        │
│ ID: ..781009 | POST | /api/v1/chat/new| 200 | 32.2s        │
├────────────────────────────────────────────────────────────┤
│ SERVER LOGS                                                │
│ ℹ️ INFO | llama runner started in 15.15 seconds            │
│ ℹ️ INFO | site.serveHTTP | status=200 | d=1.6ms            │
└────────────────────────────────────────────────────────────┘
```

## 7. Success Criteria
- [ ] Successfully fetch and display running models.
- [ ] Real-time tailing of log files without high CPU usage.
- [ ] Correct parsing and visualization of `http.d` values.
- [ ] Responsive UI that handles terminal resizing.
