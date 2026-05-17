//go:build windows
package ollama

import (
	"syscall"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	ES_CONTINUOUS       = 0x80000000
	ES_SYSTEM_REQUIRED  = 0x00000001
	ES_AWAYMODE_REQUIRED = 0x00000040
)

func preventSleepWindows() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_SYSTEM_REQUIRED | ES_AWAYMODE_REQUIRED))
}

func restoreSleepWindows() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
}
