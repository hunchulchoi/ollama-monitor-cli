package ollama

import (
	"regexp"
	"time"
)

type LogEntry struct {
	Time         time.Time
	Level        string
	Msg          string
	ResponseTime time.Duration
	RequestID    string
	Method       string
	Path         string
	Status       string
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
)

func ParseLine(line string) *LogEntry {
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

	return entry
}
