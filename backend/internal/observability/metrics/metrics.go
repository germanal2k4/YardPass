package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"yardpass/internal/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NoopRegisterer struct{}

type Metrics struct {
	enabled    bool
	wg         sync.WaitGroup
	metricsLgr *metricsErrorLogger
	server     *http.Server

	r *prometheus.Registry
}

func NewMetrics(lc fx.Lifecycle, c *config.MetricsConfig, lgr *zap.Logger) (*Metrics, error) {
	if c == nil || !c.Enabled {
		return &Metrics{enabled: false, r: nil}, nil
	}

	metrics := &Metrics{
		enabled: c.Enabled,
		r:       prometheus.NewRegistry(),
		metricsLgr: &metricsErrorLogger{
			lgr.Sugar().With("component", "metrics"),
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics.r, promhttp.HandlerOpts{
		ErrorLog: metrics.metricsLgr,
		Registry: metrics.r,
	}))

	metrics.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			metrics.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return metrics.Stop(ctx)
		},
	})

	return metrics, nil
}

func (m *Metrics) Start() {
	if !m.enabled || m.server == nil {
		return
	}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.metricsLgr.Errorf("Listen http server: %s", err.Error())
		}
	}()
}

func (m *Metrics) Stop(ctx context.Context) error {
	if !m.enabled || m.server == nil {
		return nil
	}
	done := make(chan struct{})
	var err error
	go func() {
		err = m.server.Shutdown(context.Background())
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (m *Metrics) Enabled() bool {
	return m.enabled
}

// Handler exposes Prometheus scrape on another HTTP server (e.g. Telegram webhook mux).
func (m *Metrics) Handler() http.Handler {
	if !m.enabled || m.r == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.NotFound(w, nil)
		})
	}
	return promhttp.HandlerFor(m.r, promhttp.HandlerOpts{
		ErrorLog: m.metricsLgr,
		Registry: m.r,
	})
}

func (m *Metrics) GetRegistry() prometheus.Registerer {
	if !m.enabled {
		return &NoopRegisterer{}
	}
	return m.r
}

type metricsErrorLogger struct {
	*zap.SugaredLogger
}

func (m *metricsErrorLogger) Println(v ...interface{}) {
	m.Error(v...)
}

func (n *NoopRegisterer) MustRegister(c ...prometheus.Collector) {
}

func (n *NoopRegisterer) Register(c prometheus.Collector) error {
	return nil
}

func (n *NoopRegisterer) Unregister(c prometheus.Collector) bool {
	return true
}
