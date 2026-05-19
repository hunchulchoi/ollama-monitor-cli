package ollama

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

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
}

// Improved regex to handle both general logs and request logs
var (
	timeRegex   = regexp.MustCompile(`time=([^\s]+)`)
	levelRegex  = regexp.MustCompile(`level=([^\s]+)`)
	msgRegex    = regexp.MustCompile(`msg=("(?:[^"\\]|\\.)*"|[^\s]+)`)
	durRegex    = regexp.MustCompile(`http\.d=([^\s]+)`)
	idRegex     = regexp.MustCompile(`request_id=([^\s]+)`)
	methodRegex = regexp.MustCompile(`http\.method=([^\s]+)`)
	pathRegex   = regexp.MustCompile(`http\.path=([^\s]+)`)
	statusRegex = regexp.MustCompile(`http\.status=([^\s]+)`)

	// Debug metrics regex
	promptEvalCountRegex    = regexp.MustCompile(`"prompt_eval_count":\s*(\d+)`)
	evalCountRegex          = regexp.MustCompile(`"eval_count":\s*(\d+)`)
	promptEvalDurationRegex = regexp.MustCompile(`"prompt_eval_duration":\s*(\d+)`)
	evalDurationRegex       = regexp.MustCompile(`"eval_duration":\s*(\d+)`)
	totalDurationRegex      = regexp.MustCompile(`"total_duration":\s*(\d+)`)
	loadDurationRegex       = regexp.MustCompile(`"load_duration":\s*(\d+)`)

	// GIN log format: [GIN] 2026/05/18 - 21:07:28 | 200 |    5.350416ms |       127.0.0.1 | GET      "/api/tags"
	ginRegex = regexp.MustCompile(`^\[GIN\]\s+(\d{4}/\d{2}/\d{2}\s+-\s+\d{2}:\d{2}:\d{2})\s+\|\s+(\d{3})\s+\|\s+([^\s|]+)\s+\|\s+([^\s|]+)\s+\|\s+([A-Z]+)\s+"([^"]+)"`)
)

func ParseLine(line string) *LogEntry {
	// Try GIN format first as it's very specific
	if m := ginRegex.FindStringSubmatch(line); len(m) > 6 {
		t, _ := time.Parse("2006/01/02 - 15:04:05", m[1])
		dur, _ := time.ParseDuration(strings.TrimSpace(m[3]))
		return &LogEntry{
			Time:         t,
			Level:        "INFO",
			Status:       m[2],
			ResponseTime: dur,
			Method:       m[5],
			Path:         m[6],
			Msg:          fmt.Sprintf("%s %s", m[5], m[6]),
		}
	}

	entry := &LogEntry{}

	if m := timeRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Time, _ = time.Parse(time.RFC3339, m[1])
	} else {
		return nil // Time is mandatory in this format
	}

	if m := levelRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Level = m[1]
	}

	if m := msgRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Msg = m[1]
	}

	if m := durRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.ResponseTime, _ = time.ParseDuration(m[1])
	}

	if m := idRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.RequestID = m[1]
	}

	if m := methodRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Method = m[1]
	}

	if m := pathRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Path = m[1]
	}

	if m := statusRegex.FindStringSubmatch(line); len(m) > 1 {
		entry.Status = m[1]
	}

	// Extract debug metrics
	if m := promptEvalCountRegex.FindStringSubmatch(line); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &entry.PromptEvalCount)
	}
	if m := evalCountRegex.FindStringSubmatch(line); len(m) > 1 {
		fmt.Sscanf(m[1], "%d", &entry.EvalCount)
	}
	if m := promptEvalDurationRegex.FindStringSubmatch(line); len(m) > 1 {
		var n int64
		fmt.Sscanf(m[1], "%d", &n)
		entry.PromptEvalDuration = time.Duration(n)
	}
	if m := evalDurationRegex.FindStringSubmatch(line); len(m) > 1 {
		var n int64
		fmt.Sscanf(m[1], "%d", &n)
		entry.EvalDuration = time.Duration(n)
	}
	if m := totalDurationRegex.FindStringSubmatch(line); len(m) > 1 {
		var n int64
		fmt.Sscanf(m[1], "%d", &n)
		entry.TotalDuration = time.Duration(n)
	}
	if m := loadDurationRegex.FindStringSubmatch(line); len(m) > 1 {
		var n int64
		fmt.Sscanf(m[1], "%d", &n)
		entry.LoadDuration = time.Duration(n)
	}

	return entry
}
