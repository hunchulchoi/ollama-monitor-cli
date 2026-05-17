package tui

import "strings"

func RenderSparkline(data []float64, width int) string {
	if len(data) == 0 {
		return "No data"
	}
	blocks := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	max := 0.0
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	start := 0
	if len(data) > width {
		start = len(data) - width
	}
	subset := data[start:]

	var sb strings.Builder
	for _, v := range subset {
		idx := int((v / max) * float64(len(blocks)-1))
		sb.WriteString(blocks[idx])
	}
	return sb.String()
}
