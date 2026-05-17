package ollama

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRunningModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"models": [{"name": "test-model", "size": 1000}]}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	res, err := client.GetRunningModels()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(res.Models) != 1 || res.Models[0].Name != "test-model" {
		t.Errorf("Unexpected result: %+v", res)
	}
}
