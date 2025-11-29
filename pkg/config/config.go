// Package config provides configuration management for the proxy server.
// It supports YAML configuration files and environment variable overrides.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the proxy server.
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Proxy   ProxyConfig   `mapstructure:"proxy"`
	Logging LoggingConfig `mapstructure:"logging"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Address            string        `mapstructure:"address"`
	ReadTimeout        time.Duration `mapstructure:"read_timeout"`
	WriteTimeout       time.Duration `mapstructure:"write_timeout"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	MaxConnsPerIP      int           `mapstructure:"max_conns_per_ip"`
	MaxRequestsPerConn int           `mapstructure:"max_requests_per_conn"`
}

// ProxyConfig holds proxy-specific configuration.
type ProxyConfig struct {
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ResponseTimeout time.Duration `mapstructure:"response_timeout"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MetricsConfig holds Prometheus metrics configuration.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
	Path    string `mapstructure:"path"`
}

// Load reads configuration from file and environment variables.
// Environment variables use PROXY_ prefix and underscore separators.
// Example: PROXY_SERVER_ADDRESS overrides server.address
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/proxy/")
		v.AddConfigPath("$HOME/.proxy/")
	}

	// Environment variable support
	v.SetEnvPrefix("PROXY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; use defaults and env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures default values for all settings.
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.address", ":8080")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.max_conns_per_ip", 10000)
	v.SetDefault("server.max_requests_per_conn", 0)

	// Proxy defaults
	v.SetDefault("proxy.dial_timeout", "10s")
	v.SetDefault("proxy.response_timeout", "60s")
	v.SetDefault("proxy.max_idle_conns", 1000)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("logging.output", "stdout")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.address", ":9090")
	v.SetDefault("metrics.path", "/metrics")
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server.address cannot be empty")
	}
	if c.Server.MaxConnsPerIP < 0 {
		return fmt.Errorf("server.max_conns_per_ip must be >= 0")
	}
	if c.Server.ReadTimeout < 0 {
		return fmt.Errorf("server.read_timeout must be >= 0")
	}
	if c.Server.WriteTimeout < 0 {
		return fmt.Errorf("server.write_timeout must be >= 0")
	}
	if c.Proxy.DialTimeout <= 0 {
		return fmt.Errorf("proxy.dial_timeout must be > 0")
	}
	if c.Metrics.Enabled && c.Metrics.Address == "" {
		return fmt.Errorf("metrics.address cannot be empty when metrics are enabled")
	}
	return nil
}
