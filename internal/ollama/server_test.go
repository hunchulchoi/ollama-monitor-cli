package ollama

import (
	"net"
	"strconv"
	"testing"
)

func TestIsPortInUse(t *testing.T) {
	// 1. Listen on a dynamic local port to ensure it's in use
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen on dynamic port: %v", err)
	}
	defer listener.Close()

	// 2. Resolve the port number
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to split host/port: %v", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	// 3. Verify IsPortInUse returns true for the active port
	if !IsPortInUse(port) {
		t.Errorf("Expected port %d to be reported as in use", port)
	}

	// 4. Close the listener to free the port
	listener.Close()

	// 5. Verify IsPortInUse returns false after closing (or at least it shouldn't be active)
	// Note: on some operating systems, TCP port release might have a tiny delay, 
	// but normally for localhost listener.Close() frees it immediately.
	if IsPortInUse(port) {
		t.Errorf("Expected port %d to be free after closing listener", port)
	}
}
