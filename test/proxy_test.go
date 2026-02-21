package test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yigitkonur/proxy-http-forward/pkg/config"
	"github.com/yigitkonur/proxy-http-forward/pkg/handler"
	"github.com/yigitkonur/proxy-http-forward/pkg/metrics"
	"github.com/yigitkonur/proxy-http-forward/pkg/pool"
)

// Shared metrics instance to avoid duplicate registration
var (
	testMetrics     *metrics.Metrics
	testMetricsOnce sync.Once
)

func getTestMetrics() *metrics.Metrics {
	testMetricsOnce.Do(func() {
		testMetrics = metrics.New()
	})
	return testMetrics
}

func TestConfigLoad(t *testing.T) {
	t.Run("loads defaults when no config file", func(t *testing.T) {
		cfg, err := config.Load("")
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, ":8080", cfg.Server.Address)
		assert.Equal(t, 10000, cfg.Server.MaxConnsPerIP)
		assert.Equal(t, "info", cfg.Logging.Level)
	})

	t.Run("validates configuration", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Address:       "",
				MaxConnsPerIP: -1,
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
	})
}

func TestPoolCreation(t *testing.T) {
	cfg := config.ProxyConfig{
		DialTimeout:     10 * time.Second,
		ResponseTimeout: 60 * time.Second,
		MaxIdleConns:    100,
	}

	p := pool.New(cfg)
	require.NotNil(t, p)

	// Get and put a client
	client := p.Get()
	assert.NotNil(t, client)
	p.Put(client)
}

func TestMetrics(t *testing.T) {
	m := getTestMetrics()
	require.NotNil(t, m)

	// Test recording metrics (won't panic on duplicate registration)
	m.RecordRequest("GET", "200", "http", 0.5)
	m.IncrementConnections()
	m.DecrementConnections()
	m.IncrementTunnels()
	m.DecrementTunnels()
	m.RecordError("http", "timeout")
}

func TestHandlerCreation(t *testing.T) {
	cfg := config.ProxyConfig{
		DialTimeout:     10 * time.Second,
		ResponseTimeout: 60 * time.Second,
		MaxIdleConns:    100,
	}

	p := pool.New(cfg)
	m := getTestMetrics() // Reuse shared metrics
	logger, _ := zap.NewDevelopment()

	h := handler.New(p, m, logger.Sugar(), cfg)
	require.NotNil(t, h)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: config.Config{
				Server: config.ServerConfig{
					Address:       ":8080",
					MaxConnsPerIP: 100,
				},
				Proxy: config.ProxyConfig{
					DialTimeout: 10 * time.Second,
				},
				Metrics: config.MetricsConfig{
					Enabled: true,
					Address: ":9090",
				},
			},
			wantErr: false,
		},
		{
			name: "empty address",
			cfg: config.Config{
				Server: config.ServerConfig{
					Address: "",
				},
			},
			wantErr: true,
		},
		{
			name: "negative max conns",
			cfg: config.Config{
				Server: config.ServerConfig{
					Address:       ":8080",
					MaxConnsPerIP: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
