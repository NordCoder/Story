package logger

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"net/http"
	"strconv"
	"time"
)

var (
	reqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Namespace: "wikifeed", Subsystem: "http", Name: "requests_total", Help: "Total HTTP requests"},
		[]string{"method", "path", "status"},
	)
	latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Namespace: "wikifeed", Subsystem: "http", Name: "request_duration_seconds", Help: "HTTP latency"},
		[]string{"method", "path"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(reqTotal, latency)
	prometheus.MustRegister(collectors.NewGoCollector())
}

func PromMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		elapsed := time.Since(start).Seconds()
		path := chi.RouteContext(r.Context()).RoutePattern()
		latency.WithLabelValues(r.Method, path).Observe(elapsed)
		reqTotal.WithLabelValues(r.Method, path, strconv.Itoa(ww.Status())).Inc()
	})
}
