# ntcharts 반응형 차트 통합 구현 계획 (Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 유니코드 블록 기반 스파크라인을 `ntcharts` 기반의 상세 반응형 라인 차트로 교체하여 터미널 크기에 맞게 자동 전환되도록 구현합니다.

**Architecture:** `internal/tui/model.go`에 `ntcharts.LineChart` 모델들을 통합하고, `Update` 메서드에서 실시간 데이터를 축적하며, `View` 메서드에서 터미널 높이에 따라 상세 차트 모드(Full)와 컴팩트 모드(Compact)를 분기하여 렌더링합니다.

**Tech Stack:** Go 1.21, Bubble Tea, Lip Gloss, github.com/NimbleMarkets/ntcharts

---

### Task 1: 의존성 추가 (Dependency Setup)

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: ntcharts 라이브러리 설치**
  
  Run: `go get github.com/NimbleMarkets/ntcharts@latest`
  Expected: go.mod 및 go.sum에 의존성이 정상적으로 추가됨.

- [ ] **Step 2: 의존성 정리 및 검증**
  
  Run: `go mod tidy`
  Expected: 에러 없이 완료됨.

- [ ] **Step 3: Commit**
  
  ```bash
  git add go.mod go.sum
  git commit -m "chore: add github.com/NimbleMarkets/ntcharts dependency"
  ```

---

### Task 2: TUI 모델 확장 및 차트 초기화 (Model Expansion & Initialization)

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: LineChart 필드 추가**
  
  `internal/tui/model.go`의 `Model` 구조체에 `ntcharts` 관련 필드를 추가합니다.

  ```go
  import (
      // ... 기존 import 생략
      "github.com/NimbleMarkets/ntcharts/linechart"
  )

  type Model struct {
      // ... 기존 필드 생략
      CPUChart      linechart.Model
      MemChart      linechart.Model
      LatencyChart  linechart.Model
      TPSChart      linechart.Model
  }
  ```

- [ ] **Step 2: NewModel에서 차트 초기화 구현**
  
  `NewModel` 함수에서 차트들을 초기화합니다. (기본 크기는 임시로 10x5로 설정하고 WindowSizeMsg에서 조정)

  ```go
  func NewModel(client *ollama.Client, debugMode bool) Model {
      cpuChart := linechart.New(20, 8)
      cpuChart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("42"))) // Spring Green
      
      memChart := linechart.New(20, 8)
      memChart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("39"))) // Deep Sky Blue

      latencyChart := linechart.New(40, 8)
      latencyChart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("214"))) // Orange

      tpsChart := linechart.New(20, 8)
      tpsChart.SetStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))) // Purple

      return Model{
          client:       client,
          DebugMode:    debugMode,
          ProxyChan:    make(chan *ollama.LogEntry, 10),
          CPUChart:     cpuChart,
          MemChart:     memChart,
          LatencyChart: latencyChart,
          TPSChart:     tpsChart,
      }
  }
  ```

- [ ] **Step 3: 컴파일 확인**
  
  Run: `go build main.go`
  Expected: 컴파일 성공 (실행하지는 않음)

- [ ] **Step 4: Commit**
  
  ```bash
  git add internal/tui/model.go
  git commit -m "feat: add and initialize ntcharts models in TUI Model"
  ```

---

### Task 3: 데이터 동기화 및 윈도우 크기 반응형 크기 조정 (Data Push & Resizing)

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: WindowSizeMsg에서 차트 크기 동적 조절**
  
  `Update` 메서드의 `tea.WindowSizeMsg` 분기에서 차트들의 너비와 높이를 조절하는 로직을 추가합니다.

  ```go
  case tea.WindowSizeMsg:
      m.width = msg.Width
      m.height = msg.Height
      
      // 상세 차트 너비 계산
      chartWidth := (msg.Width - 10) / 2
      if chartWidth < 10 {
          chartWidth = 10
      }
      
      m.CPUChart.Resize(chartWidth, 8)
      m.MemChart.Resize(chartWidth, 8)
      m.LatencyChart.Resize(msg.Width-6, 8)
      m.TPSChart.Resize(chartWidth, 8)
  ```

- [ ] **Step 2: 수집된 데이터를 차트 모델에 Push**
  
  `Update` 루프에서 CPU, Memory, Latency, TPS 데이터가 수집될 때 각 차트 모델에 데이터를 주입합니다.

  ```go
  // CPU/MEM 데이터 업데이트 분기 (TickMsg 처리부 내)
  if stats != nil {
      m.CPUHistory = append(m.CPUHistory, stats.CPU)
      m.CPUChart.Push(stats.CPU) // ntcharts 연동
      
      m.MemHistory = append(m.MemHistory, stats.Memory)
      m.MemChart.Push(stats.Memory / (1024 * 1024 * 1024)) // GB 단위로 변환 후 주입
  }

  // Latency 데이터 업데이트 분기 (handleLogEntry 내)
  if entry.ResponseTime > 0 {
      m.Latencies = append(m.Latencies, float64(entry.ResponseTime.Milliseconds()))
      m.LatencyChart.Push(float64(entry.ResponseTime.Milliseconds()))
  }

  // TPS 데이터 업데이트 분기 (handleLogEntry 내)
  if entry.EvalDuration > 0 {
      tps := float64(entry.EvalCount) / entry.EvalDuration.Seconds()
      m.TPSHistory = append(m.TPSHistory, tps)
      m.TPSChart.Push(tps)
  }
  ```

- [ ] **Step 3: 빌드 검증**
  
  Run: `go build main.go`
  Expected: 성공

- [ ] **Step 4: Commit**
  
  ```bash
  git add internal/tui/model.go
  git commit -m "feat: implement dynamic resizing and data push for ntcharts"
  ```

---

### Task 4: 반응형 대시보드 뷰 구현 (Adaptive View Rendering)

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: View 메서드 내 렌더링 분기 추가**
  
  터미널 높이(`m.height`)에 따라 다른 대시보드 레이아웃을 생성합니다.

  ```go
  // internal/tui/model.go: View() 내
  
  // 높이 임계값 확인 (Full Mode vs Compact Mode)
  isFullMode := m.height >= 38

  // Resources Section
  resourcesView := LatencyStyle.Render(" 📊 RESOURCE USAGE (History)") + "\n"
  if isFullMode {
      // 2열 레이아웃으로 차트 배치
      cpuRender := m.CPUChart.View()
      memRender := m.MemChart.View()
      
      resourcesView += lipgloss.JoinHorizontal(lipgloss.Top, 
          "  CPU:\n"+cpuRender, 
          "   MEM:\n"+memRender,
      )
  } else {
      // 기존 스파크라인 뷰 유지
      sparkWidth := (contentWidth - 16) / 2
      if sparkWidth < 5 {
          sparkWidth = 5
      }
      cpuSpark := RenderSparkline(m.CPUHistory, sparkWidth, 1.0)
      memSpark := RenderSparkline(m.MemHistory, sparkWidth, 1024*1024*1024)
      resourcesView += fmt.Sprintf("  CPU: %-*s  MEM: %-*s", sparkWidth, cpuSpark, sparkWidth, memSpark)
  }
  
  // Performance Section (동일하게 분기 처리)
  performanceView := LatencyStyle.Render(" ⚡ PERFORMANCE (Latency Flow)") + "\n"
  if isFullMode {
      performanceView += "  " + m.LatencyChart.View()
  } else {
      sparkline := RenderSparkline(m.Latencies, contentWidth-4, 1000.0)
      if sparkline == "No data" {
          performanceView += "  No data yet..."
      } else {
          performanceView += "  " + sparkline
      }
  }
  ```

- [ ] **Step 2: 애플리케이션 실행 및 수동 검증**
  
  Run: `go run main.go`
  Expected: TUI 대시보드가 에러 없이 정상적으로 켜지며, 창 크기(높이)를 늘리고 줄임에 따라 차트 모드가 동적으로 바뀜.

- [ ] **Step 3: Commit**
  
  ```bash
  git add internal/tui/model.go
  git commit -m "feat: complete adaptive view rendering with ntcharts"
  ```
