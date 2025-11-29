<h1 align="center">🦑 Go Native Squid Proxy 🦑</h1>
<h3 align="center">Stop babysitting Squid configs. Start proxying at warp speed.</h3>

<p align="center">
  <strong>
    <em>A high-performance HTTP/HTTPS proxy server written in pure Go. It's Squid, but actually fast, and you don't need a PhD to configure it.</em>
  </strong>
</p>

<p align="center">
  <!-- Package Info -->
  <a href="#"><img alt="go" src="https://img.shields.io/badge/go-1.21+-00ADD8.svg?style=flat-square&logo=go"></a>
  <a href="#"><img alt="fasthttp" src="https://img.shields.io/badge/powered_by-fasthttp-00ADD8.svg?style=flat-square"></a>
  &nbsp;&nbsp;•&nbsp;&nbsp;
  <!-- Features -->
  <a href="#"><img alt="license" src="https://img.shields.io/badge/License-MIT-F9A825.svg?style=flat-square"></a>
  <a href="#"><img alt="platform" src="https://img.shields.io/badge/platform-macOS_|_Linux_|_Windows-2ED573.svg?style=flat-square"></a>
</p>

<p align="center">
  <img alt="zero config" src="https://img.shields.io/badge/⚙️_zero_config-works_out_of_the_box-2ED573.svg?style=for-the-badge">
  <img alt="prometheus ready" src="https://img.shields.io/badge/📊_prometheus-built--in_metrics-2ED573.svg?style=for-the-badge">
</p>

<div align="center">

### 🧭 Quick Navigation

[**⚡ Get Started**](#-get-started-in-60-seconds) •
[**✨ Key Features**](#-feature-breakdown-the-secret-sauce) •
[**🎮 Usage & Examples**](#-usage-fire-and-forget) •
[**⚙️ Configuration**](#️-configuration) •
[**🐳 Docker**](#-docker)

</div>

---

**Go Native Squid Proxy** is the proxy server your infrastructure wishes it had. Stop wrestling with arcane Squid configurations and cryptic error messages. This proxy is built with Go's legendary concurrency, powered by fasthttp for maximum throughput, and designed to just work out of the box.

<div align="center">
<table>
<tr>
<td align="center">
<h3>⚡</h3>
<b>Blazing Fast</b><br/>
<sub>Built on fasthttp, 10x faster than net/http</sub>
</td>
<td align="center">
<h3>🔒</h3>
<b>HTTPS Tunneling</b><br/>
<sub>Full CONNECT support for TLS passthrough</sub>
</td>
<td align="center">
<h3>📊</h3>
<b>Prometheus Ready</b><br/>
<sub>Metrics endpoint out of the box</sub>
</td>
</tr>
</table>
</div>

How it slaps:
- **You:** `make run`
- **Proxy:** Starts instantly, ready to handle 10k+ connections
- **You:** `curl -x localhost:8080 https://httpbin.org/ip`
- **Result:** Your IP is proxied. Zero config. Just works.

---

## 💥 Why This Slaps Other Proxies

Setting up Squid is a nightmare. Go Native Squid Proxy makes legacy proxies look ancient.

<table align="center">
<tr>
<td align="center"><b>❌ The Squid Way (Pain)</b></td>
<td align="center"><b>✅ The Go Native Way (Glory)</b></td>
</tr>
<tr>
<td>
<ol>
  <li>Install Squid via package manager</li>
  <li>Edit 500-line squid.conf</li>
  <li>Google every ACL syntax</li>
  <li>Restart, pray, check logs</li>
  <li>Memory leaks after a week</li>
</ol>
</td>
<td>
<ol>
  <li><code>make build</code></li>
  <li><code>./build/proxy</code></li>
  <li>Done.</li>
  <li>Go grab a coffee. ☕</li>
</ol>
</td>
</tr>
</table>

We're not just another proxy. We're building a **production-ready, observable, zero-config proxy** with connection pooling, structured logging, and Prometheus metrics baked in.

---

## 🚀 Get Started in 60 Seconds

<div align="center">

| Method | One-liner |
|:------:|:----------|
| 🔨 **Build from source** | `make build && ./build/proxy` |
| 🐳 **Docker** | `docker run -p 8080:8080 yigitkonur/go-native-squid-proxy` |
| 📦 **Go install** | `go install github.com/yigitkonur/go-native-squid-proxy/cmd/proxy@latest` |

</div>

### 🔨 Build from Source

```bash
# Clone the repo
git clone https://github.com/yigitkonur/go-native-squid-proxy.git
cd go-native-squid-proxy

# Build and run
make build
./build/proxy
```

### 🐳 Docker (Recommended for Production)

```bash
# Build the image
make docker-build

# Run it
docker run -d \
  --name proxy \
  -p 8080:8080 \
  -p 9090:9090 \
  go-native-squid-proxy:latest
```

### 📦 Go Install

```bash
go install github.com/yigitkonur/go-native-squid-proxy/cmd/proxy@latest
proxy
```

> **✨ Zero Config:** The proxy starts with sensible defaults. No config file needed!

---

## 🎮 Usage: Fire and Forget

### Basic Proxying

```bash
# Start the proxy (default: :8080)
./build/proxy

# Test HTTP proxying
curl -x localhost:8080 http://httpbin.org/ip

# Test HTTPS proxying (CONNECT tunneling)
curl -x localhost:8080 https://httpbin.org/ip

# Use with any HTTP client
export http_proxy=http://localhost:8080
export https_proxy=http://localhost:8080
wget https://example.com
```

### Check Metrics

```bash
# Prometheus metrics are available at :9090/metrics
curl http://localhost:9090/metrics

# Key metrics available:
# - proxy_requests_total{method, status, type}
# - proxy_request_duration_seconds{method, type}
# - proxy_active_connections
# - proxy_tunnel_connections
# - proxy_bytes_sent_total{type}
# - proxy_bytes_received_total{type}
# - proxy_errors_total{type, reason}
```

### Command Line Options

```bash
# Show version
./build/proxy -version

# Use custom config file
./build/proxy -config /path/to/config.yaml

# Override with environment variables
PROXY_SERVER_ADDRESS=":9999" ./build/proxy
```

---

## ✨ Feature Breakdown: The Secret Sauce

<div align="center">

| Feature | What It Does | Why You Care |
| :---: | :--- | :--- |
| **⚡ fasthttp Engine**<br/>10x faster than net/http | Uses fasthttp for request handling with zero-allocation design | Handle 100k+ req/sec on modest hardware |
| **🔒 CONNECT Tunneling**<br/>Full HTTPS support | Implements HTTP CONNECT method for transparent TLS passthrough | Proxy HTTPS without breaking encryption |
| **📊 Prometheus Metrics**<br/>Observable by default | Exposes request counts, latencies, connection stats, error rates | Plug into Grafana dashboards instantly |
| **🔄 Connection Pooling**<br/>Efficient resource use | Reuses upstream connections with intelligent pooling | Reduce latency, save file descriptors |
| **📝 Structured Logging**<br/>Powered by zap | JSON or console logs with configurable levels | Debug issues fast, grep-friendly logs |
| **⚙️ Env Overrides**<br/>12-factor ready | Override any config with `PROXY_*` env vars | Perfect for containers and K8s |
| **🛡️ Graceful Shutdown**<br/>Zero dropped requests | Handles SIGTERM/SIGINT with connection draining | Zero-downtime deployments |
| **🎯 Hop-by-Hop Handling**<br/>RFC compliant | Properly strips proxy headers per HTTP spec | No header leakage, clean proxying |

</div>

---

## ⚙️ Configuration

Configuration uses YAML with environment variable overrides.

### Default Config (`config.yaml`)

```yaml
# Server settings
server:
  address: ":8080"           # Proxy listen address
  read_timeout: 30s          # Read timeout
  write_timeout: 30s         # Write timeout
  idle_timeout: 120s         # Keep-alive timeout
  max_conns_per_ip: 10000    # Max connections per IP
  max_requests_per_conn: 0   # 0 = unlimited

# Proxy behavior
proxy:
  dial_timeout: 10s          # Upstream connect timeout
  response_timeout: 60s      # Upstream response timeout
  max_idle_conns: 1000       # Pooled connections

# Logging
logging:
  level: "info"              # debug, info, warn, error
  format: "console"          # console or json
  output: "stdout"           # stdout, stderr, or file path

# Prometheus metrics
metrics:
  enabled: true
  address: ":9090"
  path: "/metrics"
```

### Environment Variable Overrides

All settings can be overridden with `PROXY_` prefix:

```bash
# Override listen address
PROXY_SERVER_ADDRESS=":9999"

# Override log level
PROXY_LOGGING_LEVEL="debug"

# Disable metrics
PROXY_METRICS_ENABLED="false"

# Example: run with overrides
PROXY_SERVER_ADDRESS=":3128" \
PROXY_LOGGING_LEVEL="debug" \
./build/proxy
```

---

## 🐳 Docker

### Quick Start

```bash
# Run with defaults
docker run -d -p 8080:8080 -p 9090:9090 go-native-squid-proxy

# Run with custom config
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -v $(pwd)/config.yaml:/etc/proxy/config.yaml \
  go-native-squid-proxy
```

### Docker Compose

```yaml
version: '3.8'

services:
  proxy:
    image: go-native-squid-proxy:latest
    ports:
      - "8080:8080"   # Proxy
      - "9090:9090"   # Metrics
    environment:
      - PROXY_LOGGING_LEVEL=info
      - PROXY_SERVER_MAX_CONNS_PER_IP=50000
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9090/metrics"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: proxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: proxy
  template:
    metadata:
      labels:
        app: proxy
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      containers:
      - name: proxy
        image: go-native-squid-proxy:latest
        ports:
        - containerPort: 8080
          name: proxy
        - containerPort: 9090
          name: metrics
        livenessProbe:
          httpGet:
            path: /metrics
            port: 9090
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 1000m
            memory: 256Mi
```

---

## 📁 Project Structure

```
go-native-squid-proxy/
├── cmd/
│   └── proxy/
│       └── main.go          # Entry point
├── pkg/
│   ├── config/              # Configuration management
│   ├── handler/             # Request handlers
│   ├── log/                 # Structured logging
│   ├── metrics/             # Prometheus metrics
│   ├── pool/                # Connection pooling
│   └── proxy/               # Server implementation
├── test/                    # Tests
├── config.yaml              # Default config
├── Dockerfile               # Multi-stage Docker build
├── Makefile                 # Build automation
└── README.md                # You are here
```

---

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make (optional but recommended)

### Build Commands

```bash
# Build binary
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Build for all platforms
make build-all

# Show all available commands
make help
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Benchmarks
make bench
```

---

## 📊 Metrics & Monitoring

### Available Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `proxy_requests_total` | Counter | method, status, type | Total requests processed |
| `proxy_request_duration_seconds` | Histogram | method, type | Request latency |
| `proxy_active_connections` | Gauge | - | Current active connections |
| `proxy_tunnel_connections` | Gauge | - | Active CONNECT tunnels |
| `proxy_bytes_sent_total` | Counter | type | Bytes sent to clients |
| `proxy_bytes_received_total` | Counter | type | Bytes received from clients |
| `proxy_errors_total` | Counter | type, reason | Error counts |

### Grafana Dashboard

Import the provided dashboard or query metrics directly:

```promql
# Request rate
rate(proxy_requests_total[5m])

# P99 latency
histogram_quantile(0.99, rate(proxy_request_duration_seconds_bucket[5m]))

# Error rate
rate(proxy_errors_total[5m]) / rate(proxy_requests_total[5m])
```

---

## 🔥 Common Issues & Quick Fixes

<details>
<summary><b>Expand for troubleshooting tips</b></summary>

| Problem | Solution |
| :--- | :--- |
| **Port already in use** | Change port with `PROXY_SERVER_ADDRESS=":9999"` or kill existing process |
| **Connection refused to upstream** | Check `proxy.dial_timeout` setting and upstream availability |
| **Too many open files** | Increase ulimit: `ulimit -n 65535` |
| **Slow responses** | Enable debug logging to identify bottlenecks: `PROXY_LOGGING_LEVEL=debug` |
| **Metrics not showing** | Ensure `metrics.enabled: true` and check `:9090/metrics` |

</details>

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📜 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgements

- **[fasthttp](https://github.com/valyala/fasthttp)** - The blazing fast HTTP engine
- **[zap](https://github.com/uber-go/zap)** - Structured logging that doesn't suck
- **[viper](https://github.com/spf13/viper)** - Configuration management made easy
- **[Prometheus](https://prometheus.io/)** - Monitoring that actually works

---

<div align="center">

**Built with 🔥 because configuring Squid is a soul-crushing waste of time.**

MIT © [Yiğit Konur](https://github.com/yigitkonur)

</div>