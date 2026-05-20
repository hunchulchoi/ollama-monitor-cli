package tui

import "testing"

func TestRenderSparkline(t *testing.T) {
	t.Run("Empty Data", func(t *testing.T) {
		res := RenderSparkline(nil, 10, 1.0)
		if res != "No data" {
			t.Errorf("Expected 'No data', got '%s'", res)
		}
	})

	t.Run("Standard Sparkline", func(t *testing.T) {
		data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		res := RenderSparkline(data, 5, 1.0)
		if len(res) == 0 || res == "No data" {
			t.Errorf("Expected rendered sparkline, got '%s'", res)
		}
	})

	t.Run("Zero and Negative Width", func(t *testing.T) {
		data := []float64{1.0, 2.0, 3.0}
		
		// This should not panic and should return an empty string
		resZero := RenderSparkline(data, 0, 1.0)
		if resZero != "" {
			t.Errorf("Expected empty string for width 0, got '%s'", resZero)
		}

		resNeg := RenderSparkline(data, -5, 1.0)
		if resNeg != "" {
			t.Errorf("Expected empty string for width -5, got '%s'", resNeg)
		}
	})
}
