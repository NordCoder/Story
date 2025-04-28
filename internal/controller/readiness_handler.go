package controller

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

// DependencyChecker описывает интерфейс для проверки зависимости.
type DependencyChecker interface {
	Ping(ctx context.Context) error
}

// ReadinessHandler обрабатывает /ready и проверяет зависимости.
type ReadinessHandler struct {
	dependencies map[string]DependencyChecker
}

// NewReadinessHandler создаёт новый ReadinessHandler.
func NewReadinessHandler() *ReadinessHandler {
	return &ReadinessHandler{
		dependencies: make(map[string]DependencyChecker),
	}
}

// AddDependency добавляет зависимость для проверки.
func (h *ReadinessHandler) AddDependency(name string, dep DependencyChecker) {
	h.dependencies[name] = dep
}

// RegisterRoutes регистрирует /ready
func (h *ReadinessHandler) RegisterRoutes(r chi.Router, path string) {
	r.Get(path, h.handleReady)
}

type dependencyStatus struct {
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}

type readinessResponse struct {
	Status       string                      `json:"status"`
	Dependencies map[string]dependencyStatus `json:"dependencies"`
}

func (h *ReadinessHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	dependencies := make(map[string]dependencyStatus)

	for name, dep := range h.dependencies {
		start := time.Now()
		err := dep.Ping(ctx)
		latency := time.Since(start).Milliseconds()

		status := "ok"
		if err != nil {
			status = "failed"
		}

		dependencies[name] = dependencyStatus{
			Status:    status,
			LatencyMS: latency,
		}
	}

	overallStatus := "ok"
	for _, d := range dependencies {
		if d.Status != "ok" {
			overallStatus = "unhealthy"
			break
		}
	}

	resp := readinessResponse{
		Status:       overallStatus,
		Dependencies: dependencies,
	}

	w.Header().Set("Content-Type", "application/json")
	if overallStatus != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(resp)
}
