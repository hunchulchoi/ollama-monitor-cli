package ollama

import (
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessStats struct {
	CPU    float64
	Memory float64 // in bytes
}

func GetProcessStats() (*ProcessStats, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(name), "ollama") {
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryInfo()
			
			stats := &ProcessStats{
				CPU: cpu,
			}
			if mem != nil {
				stats.Memory = float64(mem.RSS)
			}
			return stats, nil
		}
	}

	return nil, nil
}
