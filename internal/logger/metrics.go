package logger

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Metrics struct {
	Registry *prometheus.Registry
	ReqTotal *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
	once     sync.Once
}

// NewMetrics создаёт и настраивает собственный Registry
func NewMetrics() *Metrics {
	m := &Metrics{
		Registry: prometheus.NewRegistry(),
		ReqTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "wikifeed",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		Latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "wikifeed",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request latency",
			},
			[]string{"method", "path"},
		),
	}
	return m
}

// Init регистрирует все collectors только один раз
func (m *Metrics) Init() {
	m.once.Do(func() {
		m.Registry.MustRegister(
			m.ReqTotal,
			m.Latency,
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	})
}

// Handler возвращает HTTP-хендлер для экспорта метрик
func (m *Metrics) Handler() http.Handler {
	m.Init()
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

// Middleware для измерения latency и request count
func (m *Metrics) PromMiddleware(next http.Handler) http.Handler {
	m.Init()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		elapsed := time.Since(start).Seconds()
		path := chi.RouteContext(r.Context()).RoutePattern()
		m.Latency.WithLabelValues(r.Method, path).Observe(elapsed)
		m.ReqTotal.WithLabelValues(r.Method, path, strconv.Itoa(ww.Status())).Inc()
	})
}
