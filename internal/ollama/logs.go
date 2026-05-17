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
}

// Improved regex based on provided log samples
var logRegex = regexp.MustCompile(`time=(?P<time>[^\s]+) level=(?P<level>[^\s]+).+msg=(?P<msg>"[^"]*"|[^\s]+).+http\.d=(?P<dur>[^\s]+).+request_id=(?P<id>[^\s]+)`)

func ParseLine(line string) *LogEntry {
	match := logRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	
	entry := &LogEntry{}
	for i, name := range logRegex.SubexpNames() {
		if i != 0 && name != "" {
			val := match[i]
			switch name {
			case "time":
				entry.Time, _ = time.Parse(time.RFC3339, val)
			case "level":
				entry.Level = val
			case "msg":
				entry.Msg = val
			case "dur":
				entry.ResponseTime, _ = time.ParseDuration(val)
			case "id":
				entry.RequestID = val
			}
		}
	}
	return entry
}
