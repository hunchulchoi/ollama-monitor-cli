package ollama

import (
	"os/exec"
	"runtime"
)

var (
	caffeinateCmd *exec.Cmd
)

// PreventSleep keeps the system awake
func PreventSleep() {
	if runtime.GOOS == "darwin" {
		// -i: prevent system idle sleep
		// -s: prevent system sleep while on AC power
		caffeinateCmd = exec.Command("caffeinate", "-i", "-s")
		caffeinateCmd.Start()
	} else if runtime.GOOS == "windows" {
		preventSleepWindows()
	}
}

// RestoreSleep allows the system to sleep normally
func RestoreSleep() {
	if runtime.GOOS == "darwin" && caffeinateCmd != nil && caffeinateCmd.Process != nil {
		caffeinateCmd.Process.Kill()
	} else if runtime.GOOS == "windows" {
		restoreSleepWindows()
	}
}
