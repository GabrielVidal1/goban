package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	serverStartTime time.Time
	watcherRunning  atomic.Bool
)

func init() {
	serverStartTime = time.Now()
	watcherRunning.Store(false)
}

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	Timestamp string `json:"timestamp"`
}

// ReadyResponse is the JSON body returned by GET /ready.
type ReadyResponse struct {
	Status   bool              `json:"status"`
	Checks   []CheckResult     `json:"checks,omitempty"`
	Uptime   string            `json:"uptime"`
	Timestamp string           `json:"timestamp"`
}

// CheckResult describes the outcome of a single readiness check.
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthHandler returns basic server health information.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "ok",
		Uptime:    time.Since(serverStartTime).Round(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyHandler checks that all critical dependencies are operational.
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checks := []CheckResult{
		{
			Name:   "watcher",
			Status: "pass",
		},
	}

	if !watcherRunning.Load() {
		checks[0].Status = "fail"
		checks[0].Message = "file watcher is not running"
	}

	allPass := true
	for _, c := range checks {
		if c.Status != "pass" {
			allPass = false
			break
		}
	}

	statusCode := http.StatusOK
	if !allPass {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(ReadyResponse{
		Status:    allPass,
		Checks:    checks,
		Uptime:    time.Since(serverStartTime).Round(time.Second).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// SetWatcherRunning updates the watcher readiness flag.
func SetWatcherRunning(running bool) {
	watcherRunning.Store(running)
}
