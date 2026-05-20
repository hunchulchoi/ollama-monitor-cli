package tui

import "strings"

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
