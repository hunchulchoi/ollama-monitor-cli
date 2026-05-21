# Network Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement real-time network traffic measurement (upload/download speed and accumulated totals) captured through the Ollama proxy server and displayed elegantly in the TUI header using curated Lip Gloss styles.

**Architecture:** Extend the `ProxyServer` to act as an `http.Handler` wrapper to measure exact HTTP request/response sizes (headers + bodies) for all proxied API requests. Ingest these metrics into the TUI Model using a dedicated 1-second interval timer tick (`BandwidthTickMsg`) to calculate real-time transfer rates (B/s) and accumulate total transfers.

**Tech Stack:** Go (Golang), Bubble Tea (TUI Framework), Lip Gloss (Styling)

---

### Task 1: Extend LogEntry Model for Network Sizes

**Files:**
- Modify: `internal/ollama/logs.go`
- Test: `internal/ollama/logs_test.go`

- [ ] **Step 1: Write the failing test**

Modify `internal/ollama/logs_test.go` to verify that `RequestSize` and `ResponseSize` fields are populated correctly inside `LogEntry` when created.

```go
// Add this test function at the end of internal/ollama/logs_test.go
func TestLogEntryNetworkSizes(t *testing.T) {
	entry := &LogEntry{
		RequestSize:  2048,
		ResponseSize: 4096,
	}
	if entry.RequestSize != 2048 {
		t.Errorf("Expected RequestSize 2048, got %d", entry.RequestSize)
	}
	if entry.ResponseSize != 4096 {
		t.Errorf("Expected ResponseSize 4096, got %d", entry.ResponseSize)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/ollama -run TestLogEntryNetworkSizes`
Expected: FAIL (compilation error: LogEntry has no field RequestSize, ResponseSize)

- [ ] **Step 3: Write minimal implementation**

Modify `internal/ollama/logs.go:10-25` to add `RequestSize` and `ResponseSize` fields to the `LogEntry` struct.

```go
type LogEntry struct {
	Time               time.Time
	Level              string
	Msg                string
	ResponseTime       time.Duration
	RequestID          string
	Method             string
	Path               string
	Status             string
	PromptEvalCount    int
	EvalCount          int
	PromptEvalDuration time.Duration
	EvalDuration       time.Duration
	TotalDuration      time.Duration
	LoadDuration       time.Duration
	RequestSize        int64 // Add this
	ResponseSize       int64 // Add this
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/ollama -run TestLogEntryNetworkSizes`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ollama/logs.go internal/ollama/logs_test.go
git commit -m "feat(ollama): add RequestSize and ResponseSize to LogEntry model"
```

---

### Task 2: Implement ProxyServer Network Measurement Wrapper

**Files:**
- Modify: `internal/ollama/proxy.go`
- Test: `internal/ollama/proxy_test.go`

- [ ] **Step 1: Write the failing test**

Add `TestProxyByteMeasurement` to `internal/ollama/proxy_test.go` to test request/response size interception.

```go
// Add to internal/ollama/proxy_test.go
import (
	"bytes"
	"net/http"
	"net/http/httptest"
)

func TestProxyByteMeasurement(t *testing.T) {
	// 1. Create dummy target server
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response-body-12345")) // 19 bytes response body
	}))
	defer target.Close()

	metricsOut := make(chan *LogEntry, 10)
	proxy, err := NewProxyServer(target.URL, metricsOut)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Mock request
	req := httptest.NewRequest("POST", "/api/tags", bytes.NewBuffer([]byte("req-body-123"))) // 12 bytes request body
	req.Header.Set("X-Request-ID", "test-network-id")
	rec := httptest.NewRecorder()

	// 3. Directly call ServeHTTP to trigger traffic measurement handler
	proxy.ServeHTTP(rec, req)

	// 4. Verify metric sent to channel
	select {
	case entry := <-metricsOut:
		if entry.RequestID != "test-network-id" {
			t.Errorf("Expected RequestID 'test-network-id', got '%s'", entry.RequestID)
		}
		if entry.RequestSize <= 12 {
			t.Errorf("Expected RequestSize > 12 (body + header), got %d", entry.RequestSize)
		}
		if entry.ResponseSize <= 19 {
			t.Errorf("Expected ResponseSize > 19 (body + header), got %d", entry.ResponseSize)
		}
		if entry.Level != "METRIC" {
			t.Errorf("Expected Level 'METRIC', got '%s'", entry.Level)
		}
	default:
		t.Fatal("Expected LogEntry in metricsOut channel, but timed out")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/ollama -run TestProxyByteMeasurement`
Expected: FAIL (compilation error: proxy.ServeHTTP undefined)

- [ ] **Step 3: Write minimal implementation**

1. Create a `responseWriterWrapper` in `internal/ollama/proxy.go` that wraps `http.ResponseWriter` and supports `http.Flusher`.
2. Implement `ServeHTTP` on `*ProxyServer` to intercept request/response sizes.
3. Update `Start` function in `internal/ollama/proxy.go` to use `s` (which now implements `http.Handler`) instead of `s.Proxy`.

Modify `internal/ollama/proxy.go` with the following changes:

```diff
 // add to imports
+	"fmt"
```

```go
// Add after type ProxyServer struct
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bodySize   int64
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bodySize += int64(n)
	return n, err
}

func (rw *responseWriterWrapper) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// ServeHTTP implements http.Handler to intercept and measure network metrics
func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Calculate Request Size (Headers + Body)
	reqHeaderSize := int64(len(r.Method) + len(r.URL.RequestURI()) + len(r.Proto) + 4)
	for k, vs := range r.Header {
		reqHeaderSize += int64(len(k))
		for _, v := range vs {
			reqHeaderSize += int64(len(v))
		}
	}

	var reqBodySize int64
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			reqBodySize = int64(len(bodyBytes))
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}
	reqTotalSize := reqHeaderSize + reqBodySize

	// 2. Wrap ResponseWriter to capture output
	rw := &responseWriterWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	startTime := time.Now()

	// 3. Delegate to reverse proxy
	s.Proxy.ServeHTTP(rw, r)

	// 4. Calculate Response Size (Headers + Body)
	respHeaderSize := int64(len(r.Proto) + 15) // Approximate "HTTP/1.1 200 OK"
	for k, vs := range rw.Header() {
		respHeaderSize += int64(len(k))
		for _, v := range vs {
			respHeaderSize += int64(len(v))
		}
	}
	respTotalSize := respHeaderSize + rw.bodySize

	// 5. Send Network Metric
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = "captured-proxy"
	}

	entry := &LogEntry{
		Time:         time.Now(),
		Level:        "METRIC",
		Msg:          "Proxy network metrics",
		RequestID:    reqID,
		Method:       r.Method,
		Path:         r.URL.Path,
		Status:       fmt.Sprintf("%d", rw.statusCode),
		ResponseTime: time.Since(startTime),
		RequestSize:  reqTotalSize,
		ResponseSize: respTotalSize,
	}

	select {
	case s.MetricsOut <- entry:
	default:
	}
}
```

And update the `Start` method in `internal/ollama/proxy.go`:

```go
func (s *ProxyServer) Start(addr string) error {
	s.server = &http.Server{
		Addr:     addr,
		Handler:  s, // Update this from s.Proxy to s
		ErrorLog: log.New(io.Discard, "", 0),
	}
	return s.server.ListenAndServe()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/ollama -run TestProxyByteMeasurement`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ollama/proxy.go internal/ollama/proxy_test.go
git commit -m "feat(ollama): implement HTTP proxy handler to measure request/response network sizes"
```

---

### Task 3: Implement FormatBytes View Helper

**Files:**
- Modify: `internal/tui/view_utils.go`
- Test: `internal/tui/view_utils_test.go`

- [ ] **Step 1: Write the failing test**

Add `TestFormatBytes` inside `internal/tui/view_utils_test.go`.

```go
// Add to internal/tui/view_utils_test.go
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			res := FormatBytes(tc.input)
			if res != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, res)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/tui -run TestFormatBytes`
Expected: FAIL (compilation error: FormatBytes undefined)

- [ ] **Step 3: Write minimal implementation**

Add `FormatBytes` function to `internal/tui/view_utils.go`.

```go
// Add to internal/tui/view_utils.go
import "fmt"

func FormatBytes(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", b/float64(div), "KMGTPE"[exp])
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/tui -run TestFormatBytes`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/view_utils.go internal/tui/view_utils_test.go
git commit -m "feat(tui): add FormatBytes utility function with TDD"
```

---

### Task 4: Integrate Bandwidth Timer & Accumulator in TUI Model

**Files:**
- Modify: `internal/tui/model.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

Add `TestTUIBandwidthAccumulator` to `internal/tui/model_test.go`.

```go
// Add to internal/tui/model_test.go
func TestTUIBandwidthAccumulator(t *testing.T) {
	client := ollama.NewClient("http://localhost:11434")
	m := NewModel(client, false)

	// 1. Ingest dummy LogEntry with network sizes
	entry := &ollama.LogEntry{
		Level:        "METRIC",
		RequestSize:  1500,
		ResponseSize: 3000,
	}
	m.handleLogEntry(entry)

	if m.uploadTemp != 1500 || m.downloadTemp != 3000 {
		t.Errorf("Expected temp buffers 1500/3000, got %d/%d", m.uploadTemp, m.downloadTemp)
	}

	// 2. Perform tick update simulation
	_, _ = m.Update(BandwidthTickMsg(time.Now()))

	if m.UploadSpeed != 1500 || m.DownloadSpeed != 3000 {
		t.Errorf("Expected speeds 1500/3000, got %f/%f", m.UploadSpeed, m.DownloadSpeed)
	}
	if m.TotalUpload != 1500 || m.TotalDownload != 3000 {
		t.Errorf("Expected totals 1500/3000, got %d/%d", m.TotalUpload, m.TotalDownload)
	}
	if m.uploadTemp != 0 || m.downloadTemp != 0 {
		t.Errorf("Expected temp buffers to be reset, got %d/%d", m.uploadTemp, m.downloadTemp)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/tui -run TestTUIBandwidthAccumulator`
Expected: FAIL (compilation error: BandwidthTickMsg and model properties undefined)

- [ ] **Step 3: Write minimal implementation**

1. Define `BandwidthTickMsg` and `doBandwidthTick()` command generator in `internal/tui/model.go`.
2. Add network bandwidth tracking fields to the `Model` struct.
3. Update `Init()` to launch the bandwidth tick.
4. Update `Update()` to process `BandwidthTickMsg` by calculating speeds and resetting temp variables.
5. Ingest request/response sizes into `uploadTemp`/`downloadTemp` inside `handleLogEntry()`.

Modify `internal/tui/model.go` as follows:

Add message and command:
```go
type BandwidthTickMsg time.Time

func doBandwidthTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return BandwidthTickMsg(t)
	})
}
```

Add fields to `Model` struct:
```go
type Model struct {
    // ... existing fields ...
	TotalUpload      int64
	TotalDownload    int64
	UploadSpeed      float64
	DownloadSpeed    float64
	uploadTemp       int64
	downloadTemp     int64
}
```

Update `Init()`:
```go
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		doTick(),
		doBandwidthTick(), // Add this
		m.tailLogFile("server.log", -1),
		m.tailLogFile("app.log", -1),
		m.waitForProxyMetrics(),
	)
}
```

Update `Update()` switch for msg types:
```go
	case BandwidthTickMsg:
		m.UploadSpeed = float64(m.uploadTemp)
		m.DownloadSpeed = float64(m.downloadTemp)
		m.TotalUpload += m.uploadTemp
		m.TotalDownload += m.downloadTemp
		m.uploadTemp = 0
		m.downloadTemp = 0
		return m, doBandwidthTick()
```

Update `handleLogEntry()`:
```go
func (m *Model) handleLogEntry(entry *ollama.LogEntry) {
	if entry == nil {
		return
	}

	if entry.RequestSize > 0 {
		m.uploadTemp += entry.RequestSize
	}
	if entry.ResponseSize > 0 {
		m.downloadTemp += entry.ResponseSize
	}
    
    // ... rest of the existing code ...
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/tui -run TestTUIBandwidthAccumulator`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/model_test.go
git commit -m "feat(tui): integrate bandwidth calculation and stats accumulation into Model"
```

---

### Task 5: Design and Render Network Metrics in Header

**Files:**
- Modify: `internal/tui/header_view.go`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

Add `TestRenderHeaderNetworkMetrics` to `internal/tui/model_test.go`.

```go
// Add to internal/tui/model_test.go
func TestRenderHeaderNetworkMetrics(t *testing.T) {
	client := ollama.NewClient("http://localhost:11434")
	m := NewModel(client, false)
	m.ProxyMode = true
	m.UploadSpeed = 1024
	m.DownloadSpeed = 2048 * 1024
	m.TotalUpload = 4096
	m.TotalDownload = 8192 * 1024

	output := m.renderHeader()

	if !strings.Contains(output, "🛜") {
		t.Errorf("Expected header to contain network emoji 🛜, got: %s", output)
	}
	if !strings.Contains(output, "1.0 KB/s") {
		t.Errorf("Expected header to contain upload speed, got: %s", output)
	}
	if !strings.Contains(output, "2.0 MB/s") {
		t.Errorf("Expected header to contain download speed, got: %s", output)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/tui -run TestRenderHeaderNetworkMetrics`
Expected: FAIL (compilation error or string mismatch)

- [ ] **Step 3: Write minimal implementation**

Modify `internal/tui/header_view.go` to append network metrics to the header when `m.ProxyMode` is active. Use Lip Gloss styles with orange accent for upload and sky blue for download.

```go
func (m *Model) renderHeader() string {
	header := HeaderStyle.Render(" 🦙 OLLAMA MONITOR") + "  " + time.Now().Format("15:04:05")
	if m.Stats != nil {
		header += fmt.Sprintf(" | CPU: %.1f%% | MEM: %.1fGB", m.Stats.CPU, m.Stats.Memory/(1024*1024*1024))
	}
	if m.DebugMode {
		header += " | " + ErrorStyle.Bold(true).Render("DEBUG ON")
	}
	if m.ProxyMode {
		// Define custom colors for network metrics
		upColor := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))   // Orange Accent
		downColor := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))  // Sky Blue Accent
		grayColor := lipgloss.NewStyle().Foreground(lipgloss.Color("244")) // Muted Gray for totals

		upSpeedStr := FormatBytes(m.UploadSpeed) + "/s"
		downSpeedStr := FormatBytes(m.DownloadSpeed) + "/s"
		upTotalStr := FormatBytes(float64(m.TotalUpload))
		downTotalStr := FormatBytes(float64(m.TotalDownload))

		networkPart := fmt.Sprintf(" | 🛜  %s %s %s | %s %s %s",
			upColor.Render("▲"), upSpeedStr, grayColor.Render("("+upTotalStr+")"),
			downColor.Render("▼"), downSpeedStr, grayColor.Render("("+downTotalStr+")"),
		)
		header += networkPart
	}
	if m.ShutdownActive {
		minutes := int(m.ShutdownDuration.Minutes())
		seconds := int(m.ShutdownDuration.Seconds()) % 60
		header += fmt.Sprintf(" | " + ErrorStyle.Bold(true).Render("SHUTDOWN IN %02d:%02d"), minutes, seconds)
	}
	return header
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/tui -run TestRenderHeaderNetworkMetrics`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/header_view.go internal/tui/model_test.go
git commit -m "feat(tui): beautifully render real-time network traffic speeds and totals in TUI header"
```

---

### Task 6: Final Verification

- [ ] **Step 1: Run all workspace tests**

Run: `go test -v ./...`
Expected: PASS for all tests.

- [ ] **Step 2: Commit and Tag**

```bash
git add .
git commit -m "chore: complete network metrics implementation with TDD"
```
