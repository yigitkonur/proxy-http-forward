high-performance HTTP/HTTPS forward proxy in pure Go. built on fasthttp, not `net/http`. handles plain HTTP forwarding and CONNECT tunneling with connection pooling, DNS caching, and Prometheus metrics out of the box.

```bash
./proxy -config config.yaml
```

point your browser or `HTTP_PROXY` at `:8080` and you're done.

[![go](https://img.shields.io/badge/go-1.21+-93450a.svg?style=flat-square)](https://go.dev/)
[![fasthttp](https://img.shields.io/badge/fasthttp-v1.55-93450a.svg?style=flat-square)](https://github.com/valyala/fasthttp)
[![license](https://img.shields.io/badge/license-MIT-grey.svg?style=flat-square)](https://opensource.org/licenses/MIT)

---

## what it does

- **HTTP forwarding** — receives `GET http://example.com/path`, strips hop-by-hop headers, chains `X-Forwarded-For`, forwards via pooled fasthttp client
- **HTTPS CONNECT tunneling** — receives `CONNECT example.com:443`, dials upstream, hijacks the connection, copies bytes bidirectionally. no TLS inspection, true opaque tunnel
- **connection pooling** — `sync.Pool` of `fasthttp.Client` instances with configurable `MaxConnsPerHost` and 5-min idle duration
- **DNS caching** — 1-hour TTL via `fasthttp.TCPDialer`, 4096 concurrent dials
- **Prometheus metrics** — separate `net/http` server so scraping never touches proxy traffic. request counters, latency histograms, active connections, byte accounting, tunnel gauges
- **structured logging** — `zap` with console (colored) or JSON output, configurable level
- **graceful shutdown** — catches `SIGINT`/`SIGTERM`, 30-second drain deadline
- **no fingerprinting** — no `Server` header, no `Date` header, header casing preserved as-is

## install

```bash
git clone https://github.com/yigitkonur/go-http-proxy-server.git
cd go-http-proxy-server
make build
```

binary lands in `./build/proxy`.

### cross-compile

```bash
make build-linux     # ./dist/proxy-linux-amd64
make build-darwin    # ./dist/proxy-darwin-arm64
make build-all
```

### docker

```bash
make docker-build && make docker-run
```

multi-stage alpine image, runs as non-root `proxy` user, exposes 8080 + 9090, built-in health check.

## usage

```bash
# with config file
./proxy -config config.yaml

# with defaults (searches ./config.yaml, /etc/proxy/config.yaml, ~/.proxy/config.yaml)
./proxy

# print version
./proxy -version

# override via env
PROXY_SERVER_ADDRESS=":3128" PROXY_LOGGING_LEVEL="debug" ./proxy
```

### test it

```bash
# HTTP
curl -x http://localhost:8080 http://httpbin.org/ip

# HTTPS
curl -x http://localhost:8080 https://httpbin.org/ip

# metrics
curl http://localhost:9090/metrics
```

## configuration

YAML config file with env var overrides (`PROXY_<SECTION>_<KEY>`). env vars take priority.

```yaml
server:
  address: ":8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_conns_per_ip: 10000
  max_requests_per_conn: 0

proxy:
  dial_timeout: 10s
  response_timeout: 60s
  max_idle_conns: 1000

logging:
  level: "info"            # debug | info | warn | error | fatal
  format: "console"        # console | json
  output: "stdout"         # stdout | stderr | /path/to/file

metrics:
  enabled: true
  address: ":9090"
  path: "/metrics"
```

### env var examples

| env var | overrides |
|:---|:---|
| `PROXY_SERVER_ADDRESS` | `server.address` |
| `PROXY_PROXY_DIAL_TIMEOUT` | `proxy.dial_timeout` |
| `PROXY_LOGGING_LEVEL` | `logging.level` |
| `PROXY_METRICS_ENABLED` | `metrics.enabled` |

## metrics

all under the `proxy` namespace:

| metric | type | labels |
|:---|:---|:---|
| `proxy_requests_total` | counter | `method`, `status`, `type` |
| `proxy_request_duration_seconds` | histogram | `method`, `type` |
| `proxy_active_connections` | gauge | — |
| `proxy_bytes_sent_total` | counter | `type` |
| `proxy_bytes_received_total` | counter | `type` |
| `proxy_errors_total` | counter | `type`, `reason` |
| `proxy_tunnel_connections` | gauge | — |

## project structure

```
cmd/proxy/
  main.go             — entry point, signal handling, graceful shutdown
pkg/
  config/config.go    — viper-based config with YAML + env var loading
  handler/handler.go  — HTTP forwarding, CONNECT tunneling, header stripping
  log/log.go          — zap logger construction
  metrics/metrics.go  — Prometheus metric definitions + separate HTTP server
  pool/pool.go        — sync.Pool of fasthttp.Client instances
  proxy/proxy.go      — server wiring, start/shutdown orchestration
test/
  proxy_test.go       — unit tests
```

## what it doesn't do

no authentication, no ACLs, no URL filtering, no content inspection, no TLS on the listener itself. it's a fast, dumb pipe. if you need those things, put it behind something that does.

## license

MIT
