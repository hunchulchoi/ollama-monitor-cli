package tui

import (
	"fmt"
	"math"
	"strings"
)

func RenderSparkline(data []float64, width int, minMax float64) string {
	if width <= 0 {
		return ""
	}
	if len(data) == 0 {
		return "No data"
	}
	blocks := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	
	// Find max in the current data
	max := 0.0
	for _, v := range data {
		if v > max {
			max = v
		}
	}

	// Use provided minMax as a floor to prevent tiny noise from filling the graph
	if max < minMax {
		max = minMax
	}

	start := 0
	if len(data) > width {
		start = len(data) - width
	}
	subset := data[start:]

	var sb strings.Builder
	for _, v := range subset {
		val := v
		if val > max {
			val = max
		}
		idx := int((val / max) * float64(len(blocks)-1))
		sb.WriteString(blocks[idx])
	}
	return sb.String()
}

func FormatBytes(b float64) string {
	if math.IsNaN(b) {
		return "NaN B"
	}
	if math.IsInf(b, 1) {
		return "+Inf B"
	}
	if math.IsInf(b, -1) {
		return "-Inf B"
	}

	sign := ""
	if b < 0 {
		sign = "-"
		b = math.Abs(b)
	}

	const unit = 1024.0
	if b < unit {
		return fmt.Sprintf("%s%.0f B", sign, b)
	}

	suffixes := []string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	div := unit
	exp := 0
	for n := b / unit; n >= unit && exp < len(suffixes)-1; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%s%.1f %s", sign, b/div, suffixes[exp])
}

