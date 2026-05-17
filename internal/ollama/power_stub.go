//go:build !windows
package ollama

func preventSleepWindows() {}
func restoreSleepWindows() {}
