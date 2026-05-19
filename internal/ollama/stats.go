package ollama

import (
	"os"
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

	myPid := int32(os.Getpid())

	for _, p := range procs {
		if p.Pid == myPid {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		lowerName := strings.ToLower(name)
		// Match 'ollama' but exclude our own binary name 'ollama-monitor'
		if strings.Contains(lowerName, "ollama") && !strings.Contains(lowerName, "monitor") {
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
