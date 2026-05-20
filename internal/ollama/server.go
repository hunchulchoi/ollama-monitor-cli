package ollama

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// RestartOllama kills any running ollama process and starts a new one
func RestartOllama(debug bool, proxy bool) error {
	// 1. Kill existing process
	if err := KillOllama(); err != nil {
		// It's okay if it fails (might not be running)
	}

	// Wait a bit for the port to be freed
	time.Sleep(2 * time.Second)

	// 2. Determine log path
	var logPath string
	customLogDir := os.Getenv("OLLAMA_LOG_DIR")
	if customLogDir != "" {
		logPath = filepath.Join(customLogDir, "server.log")
	} else if runtime.GOOS == "windows" {
		logPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "Ollama", "server.log")
	} else {
		home, _ := os.UserHomeDir()
		logPath = filepath.Join(home, ".ollama", "logs", "server.log")
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(logPath), 0755)

	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// 3. Start new process
	cmd := exec.Command("ollama", "serve")
	
	// Set environment variables
	env := os.Environ()
	
	// Remove any existing OLLAMA_DEBUG/OLLAMA_HOST from environment to prevent conflicts
	newEnv := []string{}
	for _, e := range env {
		if !strings.HasPrefix(e, "OLLAMA_DEBUG=") && !strings.HasPrefix(e, "OLLAMA_HOST=") {
			newEnv = append(newEnv, e)
		}
	}
	
	// Add our specific settings at the end
	if debug {
		newEnv = append(newEnv, "OLLAMA_DEBUG=1")
	}
	if proxy {
		// Set OLLAMA_HOST to 11435 so the monitor can listen on 11434 and proxy
		newEnv = append(newEnv, "OLLAMA_HOST=127.0.0.1:11435")
	}
	cmd.Env = newEnv

	// Redirect output to file
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Detach the process so it keeps running
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start ollama: %v", err)
	}

	return nil
}

// KillOllama kills all running ollama processes and ensures port 11434 is free
func KillOllama() error {
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/IM", "ollama.exe", "/T").Run()
	} else {
		// Kill by process name
		exec.Command("pkill", "-9", "-x", "ollama").Run()
		exec.Command("pkill", "-9", "-x", "Ollama").Run()
		exec.Command("pkill", "-9", "-f", "ollama serve").Run()

		// Kill whatever is holding port 11434 (the most common culprit)
		cmd := exec.Command("sh", "-c", "lsof -ti:11434 | xargs kill -9")
		cmd.Run()
	}

	// Wait up to 5 seconds for the port to actually clear
	for i := 0; i < 10; i++ {
		if !isPortInUse(11434) {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("port 11434 is still in use after kill attempts")
}

func isPortInUse(port int) bool {
	// Simple check by trying to listen on the port
	// but since we are a CLI, it's easier to check via netstat/lsof for robustness
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("netstat", "-ano")
	} else {
		cmd = exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	}
	output, _ := cmd.CombinedOutput()
	return len(output) > 0
}

// IsDebugMode checks if the current running ollama process has OLLAMA_DEBUG=1
func IsDebugMode() bool {
	procs, err := process.Processes()
	if err != nil {
		return false
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
		if strings.Contains(lowerName, "ollama") && !strings.Contains(lowerName, "monitor") {
			// Get environment variables
			environ, err := p.Environ()
			if err != nil {
				continue
			}

			for _, env := range environ {
				if env == "OLLAMA_DEBUG=1" || env == "OLLAMA_DEBUG=2" {
					return true
				}
			}
		}
	}

	return false
}
