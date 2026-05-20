package ollama

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// ProxyMetrics stores the captured token info
type ProxyMetrics struct {
	Model              string        `json:"model"`
	EvalCount          int           `json:"eval_count"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	EvalDuration       time.Duration `json:"eval_duration"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration"`
	TotalDuration      time.Duration `json:"total_duration"`
}

type ProxyServer struct {
	TargetURL  *url.URL
	Proxy      *httputil.ReverseProxy
	MetricsOut chan *LogEntry
	server     *http.Server
}

func NewProxyServer(target string, metricsOut chan *LogEntry) (*ProxyServer, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	s := &ProxyServer{
		TargetURL:  u,
		MetricsOut: metricsOut,
	}

	s.Proxy = httputil.NewSingleHostReverseProxy(u)
	s.Proxy.ModifyResponse = s.modifyResponse
	s.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Send the proxy error to TUI so it shows up in SERVER LOGS nicely
		entry := &LogEntry{
			Time:         time.Now(),
			Level:        "ERROR",
			Msg:          "Proxy error: " + err.Error(),
			RequestID:    r.Header.Get("X-Request-ID"),
			Method:       r.Method,
			Path:         r.URL.Path,
			Status:       "502",
			ResponseTime: 0,
		}
		if entry.RequestID == "" {
			entry.RequestID = "captured-proxy"
		}
		select {
		case s.MetricsOut <- entry:
		default:
		}
		w.WriteHeader(http.StatusBadGateway)
	}
	
	return s, nil
}

func (s *ProxyServer) Start(addr string) error {
	s.server = &http.Server{
		Addr:     addr,
		Handler:  s.Proxy,
		ErrorLog: log.New(io.Discard, "", 0), // Discard default server logging to prevent screen clutter
	}
	return s.server.ListenAndServe()
}

func (s *ProxyServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *ProxyServer) modifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	path := resp.Request.URL.Path
	if path != "/api/generate" && path != "/api/chat" {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	method := resp.Request.Method
	reqID := resp.Request.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = "captured-proxy"
	}

	var metrics ProxyMetrics
	if err := json.Unmarshal(body, &metrics); err == nil && metrics.EvalCount > 0 {
		s.sendMetrics(&metrics, method, path, reqID)
		return nil
	}

	lines := bytes.Split(body, []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		if len(lines[i]) == 0 {
			continue
		}
		var m ProxyMetrics
		if err := json.Unmarshal(lines[i], &m); err == nil && m.EvalCount > 0 {
			s.sendMetrics(&m, method, path, reqID)
			break
		}
	}

	return nil
}

func (s *ProxyServer) sendMetrics(m *ProxyMetrics, method, path, reqID string) {
	entry := &LogEntry{
		Time:               time.Now(),
		Level:              "METRIC",
		Msg:                "Finish generation (captured via proxy)",
		EvalCount:          m.EvalCount,
		PromptEvalCount:    m.PromptEvalCount,
		EvalDuration:       m.EvalDuration,
		PromptEvalDuration: m.PromptEvalDuration,
		TotalDuration:      m.TotalDuration,
		ResponseTime:       m.TotalDuration, // Map TotalDuration to ResponseTime
		RequestID:          reqID,
		Method:             method,
		Path:               path,
		Status:             "200",
	}
	
	select {
	case s.MetricsOut <- entry:
	default:
	}
}
