# [Spec] Real-Time Network Bandwidth & Stats Visualization Charts

This document specifies the design for integrating real-time network bandwidth (upload and download speeds) into the existing Bubbletea-based terminal dashboard as beautiful visualization charts and sparklines.

## 1. Goal & Background

Currently, the Ollama Monitor CLI measures live network throughput (upload/download speeds and total bytes transferred) strictly via the local proxy, but displays it only as text metrics in the header. To provide a premium, state-of-the-art terminal monitoring experience, we will visualize this dynamic bandwidth telemetry in real-time.

Choosing the approved **Approach A**, we will expand the `⚡ PERFORMANCE` view into a unified `⚡ NETWORK & PERFORMANCE` view. When `ProxyMode` is active, it will present three side-by-side or stacked real-time charts:
1. **API Latency Flow** (Latency Chart - Orange)
2. **Upload Speed Flow** (Upload Chart - Magenta/Purple)
3. **Download Speed Flow** (Download Chart - Cyan/Blue)

This design complies with our core values: **SOLID architectural separation**, **fail-safe formatting**, and **Obsidian-optimized documentation**.

---

## 2. Proposed UI/UX Layout Design

### A. Full Terminal Mode (`isFullMode == true`, Height >= 38)
We will divide the container into a 3-column side-by-side layout using `lipgloss.JoinHorizontal` to maximize horizontal space:

```
┌─────────────────────────────────────────────── NETWORK & PERFORMANCE ────────────────────────────────────────────────┐
│  [ Latency Flow ]                      [ Upload Speed Flow ]                  [ Download Speed Flow ]                │
│  800ms ┐       .                       5.2MB/s ┐       .                      12.4MB/s ┐        .                    │
│  400ms │   .  / \                      2.6MB/s │   .  / \                     6.2 MB/s │    .  / \                   │
│    0ms └─┴─┴─┴─┴─┴─                    0.0 B/s └─┴─┴─┴─┴─┴─                    0.0  B/s └─┴─┴─┴─┴─┴─                  │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

* **Color Palette (Lip Gloss tailored)**:
  * **Latency**: Warm Orange (`lipgloss.Color("214")`)
  * **Upload**: Vivid Purple/Magenta (`lipgloss.Color("170")` or `"5"`)
  * **Download**: Bright Cyan/Blue (`lipgloss.Color("51")` or `"39"`)

### B. Compact Terminal Mode (`isFullMode == false`, Height < 38)
When vertical space is limited, the charts collapse into three sleek, labeled sparklines to prevent vertical bloat:

```
⚡ NETWORK & PERFORMANCE
  Latency:  [ ▂▃▅▆▇  ] (Latest: 230ms)
  Upload:   ↑ [▂▃    ] (Latest: 12.4 KB/s | Total: 45.2 MB)
  Download: ↓ [▂▃▅▆▇█] (Latest: 2.1 MB/s | Total: 1.2 GB)
```

---

## 3. Architecture & Data Flow

### A. Model States Extension (`internal/tui/model.go`)
We will add two new streamlined charts and historical bandwidth lists to the `Model` struct:

```go
type Model struct {
    ...
    UploadChart      linechart.Model
    DownloadChart    linechart.Model
    UploadHistory    []float64 // KB/s
    DownloadHistory  []float64 // KB/s
}
```

### B. Ingesting Bandwidth Metrics on Tick
In the `Update()` message loop, under `BandwidthTickMsg`, we compute the instant speed rates and feed them directly into the stream charts:

1. Convert raw byte speeds (`uploadTemp`, `downloadTemp`) into **KB/s** (dividing by `1024.0`) to keep the chart values easily scalable.
2. Push the calculated speeds to `UploadChart` and `DownloadChart`.
3. Append values to `UploadHistory` and `DownloadHistory` (keeping a rolling cap of `50` entries for compact sparkline rendering).

```go
case BandwidthTickMsg:
    m.UploadSpeed = float64(m.uploadTemp)
    m.DownloadSpeed = float64(m.downloadTemp)
    m.TotalUpload += m.uploadTemp
    m.TotalDownload += m.downloadTemp
    
    // Feed charts (in KB/s)
    upKB := m.UploadSpeed / 1024.0
    downKB := m.DownloadSpeed / 1024.0
    
    m.UploadChart.Push(upKB)
    m.DownloadChart.Push(downKB)
    
    m.UploadHistory = append(m.UploadHistory, upKB)
    m.DownloadHistory = append(m.DownloadHistory, downKB)
    
    // Cap histories at 50 to avoid memory growth
    if len(m.UploadHistory) > 50 {
        m.UploadHistory = m.UploadHistory[1:]
    }
    if len(m.DownloadHistory) > 50 {
        m.DownloadHistory = m.DownloadHistory[1:]
    }
    
    m.uploadTemp = 0
    m.downloadTemp = 0
    return m, doBandwidthTick()
```

### C. Responsive Window Resizing
Under `tea.WindowSizeMsg`, compute a 3-way split of the viewport width. Apply safety floors to ensure layout calculations never panic.

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height

    // 2-way split for CPU / Memory
    chartWidth := (msg.Width - 10) / 2
    if chartWidth < 10 { chartWidth = 10 }
    m.CPUChart.Resize(chartWidth, 8)
    m.MemChart.Resize(chartWidth, 8)
    m.TPSChart.Resize(chartWidth, 8)

    // 3-way split for Latency, Upload, and Download in the Expanded Performance section
    thirdWidth := (msg.Width - 12) / 3
    if thirdWidth < 10 { thirdWidth = 10 }
    
    m.LatencyChart.Resize(thirdWidth, 8)
    m.UploadChart.Resize(thirdWidth, 8)
    m.DownloadChart.Resize(thirdWidth, 8)
```

### D. Layout View Composition (`internal/tui/performance_view.go`)
We will rewrite `renderPerformance(...)` to assemble the multi-column layout using Lip Gloss side-by-side borders:

* For **Full Mode**:
  1. Draw `LatencyChart`, `UploadChart`, and `DownloadChart`.
  2. Sub-render titles for each column (e.g. `Latency Flow`, `Upload (KB/s)`, `Download (KB/s)`).
  3. Combine columns horizontally using `lipgloss.JoinHorizontal(lipgloss.Top, latCol, upCol, downCol)`.
* For **Compact Mode**:
  1. Render three clean bullet points containing labeled sparklines using `RenderSparkline`.
  2. Integrate formatted speed metrics and totals.

---

## 4. Robustness & Edge Cases (Fail-safe Formatter)

1. **Quiet Traffic Guard**: When there is no active traffic (`0 B/s`), the chart axis must not crash or display incorrect infinite values. The charts will automatically handle flat zero lines cleanly.
2. **Speed Formatters**: Re-use the robust `FormatBytes` utility in `view_utils.go` (which guards against `NaN` and `Inf`) to suffix metrics correctly as `B/s`, `KB/s`, `MB/s`, etc.
3. **Responsive Width Buffer**: Ensure the 3-column width divisor handles margins properly so that Bubbletea does not wrap lines when resizing the window.

---

## 5. Verification & Testing Plan

### A. Automated Tests
We will add unit tests in `internal/tui/model_test.go`:
* **TestNetworkChartIngestion**: Asset that `BandwidthTickMsg` correctly updates upload/download history and pushes entries to `UploadChart`/`DownloadChart`.
* **TestNetworkChartResize**: Verify that sending a `tea.WindowSizeMsg` correctly resizes `UploadChart` and `DownloadChart` without causing zero-division or out-of-bounds panics.

Run command:
```bash
go test -v ./internal/tui -run "TestNetworkChart"
```

### B. Manual Verification
1. Run the dashboard using `go run main.go`.
2. Press `p` to enable Proxy Mode (listens on `11435`).
3. Send high-throughput requests to Ollama through the proxy (e.g. `curl http://localhost:11435/api/tags` or run generation prompts).
4. Verify charts dynamically scale up and down on the terminal.
5. Resize the terminal window to confirm the 3-way layout remains responsive.
