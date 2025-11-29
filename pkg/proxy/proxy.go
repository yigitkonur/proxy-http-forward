// Package proxy provides the main proxy server implementation.
package proxy

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/yigitkonur/go-native-squid-proxy/pkg/config"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/handler"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/metrics"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/pool"
)

// Server represents the proxy server.
type Server struct {
	config        *config.Config
	logger        *zap.SugaredLogger
	server        *fasthttp.Server
	metricsServer *metrics.Server
	handler       *handler.Handler
	pool          *pool.Pool
	metrics       *metrics.Metrics
}

// New creates a new proxy server.
func New(cfg *config.Config, logger *zap.SugaredLogger) *Server {
	// Initialize metrics
	m := metrics.New()

	// Initialize connection pool
	p := pool.New(cfg.Proxy)

	// Initialize handler
	h := handler.New(p, m, logger, cfg.Proxy)

	// Create fasthttp server
	server := &fasthttp.Server{
		Handler:               h.HandleRequest,
		Name:                  "go-native-squid-proxy",
		ReadTimeout:          cfg.Server.ReadTimeout,
		WriteTimeout:         cfg.Server.WriteTimeout,
		IdleTimeout:          cfg.Server.IdleTimeout,
		MaxConnsPerIP:        cfg.Server.MaxConnsPerIP,
		MaxRequestsPerConn:   cfg.Server.MaxRequestsPerConn,
		DisableKeepalive:     false,
		TCPKeepalive:         true,
		TCPKeepalivePeriod:   time.Minute,
		NoDefaultServerHeader: true,
		NoDefaultDate:        true,
		DisableHeaderNamesNormalizing: true,
	}

	s := &Server{
		config:  cfg,
		logger:  logger,
		server:  server,
		pool:    p,
		metrics: m,
		handler: h,
	}

	// Initialize metrics server if enabled
	if cfg.Metrics.Enabled {
		s.metricsServer = metrics.NewServer(cfg.Metrics)
	}

	return s
}

// Start starts the proxy server.
func (s *Server) Start() error {
	// Start metrics server if enabled
	if s.metricsServer != nil {
		go func() {
			s.logger.Infow("starting metrics server",
				"address", s.config.Metrics.Address,
				"path", s.config.Metrics.Path,
			)
			if err := s.metricsServer.Start(); err != nil {
				s.logger.Errorw("metrics server error", "error", err)
			}
		}()
	}

	s.logger.Infow("starting proxy server",
		"address", s.config.Server.Address,
		"max_conns_per_ip", s.config.Server.MaxConnsPerIP,
	)

	return s.server.ListenAndServe(s.config.Server.Address)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down proxy server...")

	// Shutdown metrics server
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(); err != nil {
			s.logger.Warnw("error shutting down metrics server", "error", err)
		}
	}

	// Shutdown main server
	return s.server.Shutdown()
}

// ShutdownWithContext gracefully shuts down the server with context.
func (s *Server) ShutdownWithContext(ctx context.Context) error {
	s.logger.Info("shutting down proxy server...")

	// Shutdown metrics server
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(); err != nil {
			s.logger.Warnw("error shutting down metrics server", "error", err)
		}
	}

	// Shutdown main server with context
	done := make(chan error, 1)
	go func() {
		done <- s.server.Shutdown()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
