// Package handler provides HTTP request handling for the proxy server.
package handler

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/yigitkonur/go-native-squid-proxy/pkg/config"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/metrics"
	"github.com/yigitkonur/go-native-squid-proxy/pkg/pool"
)

// hopByHopHeaders lists headers that should not be forwarded.
var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// Handler handles HTTP proxy requests.
type Handler struct {
	pool    *pool.Pool
	metrics *metrics.Metrics
	logger  *zap.SugaredLogger
	config  config.ProxyConfig
}

// New creates a new Handler.
func New(p *pool.Pool, m *metrics.Metrics, logger *zap.SugaredLogger, cfg config.ProxyConfig) *Handler {
	return &Handler{
		pool:    p,
		metrics: m,
		logger:  logger,
		config:  cfg,
	}
}

// HandleRequest is the main request handler for the proxy.
func (h *Handler) HandleRequest(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	h.metrics.IncrementConnections()
	defer h.metrics.DecrementConnections()

	method := string(ctx.Method())

	// Handle HTTP CONNECT method for HTTPS tunneling
	if method == fasthttp.MethodConnect {
		h.handleConnect(ctx, start)
		return
	}

	// Handle regular HTTP proxy requests
	h.handleHTTP(ctx, start)
}

// handleHTTP proxies regular HTTP requests.
func (h *Handler) handleHTTP(ctx *fasthttp.RequestCtx, start time.Time) {
	method := string(ctx.Method())

	// Prepare the outgoing request
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// Copy request from context
	ctx.Request.CopyTo(req)

	// Remove hop-by-hop headers
	removeHopByHopHeaders(&req.Header)

	// Add X-Forwarded-For header
	clientIP := ctx.RemoteIP().String()
	if xff := string(req.Header.Peek("X-Forwarded-For")); xff != "" {
		req.Header.Set("X-Forwarded-For", xff+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	// Execute the request
	err := h.pool.DoTimeout(req, resp, h.config.ResponseTimeout)
	if err != nil {
		h.handleError(ctx, start, method, "http", err, "upstream_request_failed")
		return
	}

	// Copy response back to context
	resp.CopyTo(&ctx.Response)

	// Remove hop-by-hop headers from response
	removeResponseHopByHopHeaders(&ctx.Response.Header)

	// Record metrics
	duration := time.Since(start).Seconds()
	status := strconv.Itoa(resp.StatusCode())
	h.metrics.RecordRequest(method, status, "http", duration)
	h.metrics.BytesSent.WithLabelValues("http").Add(float64(len(resp.Body())))
	h.metrics.BytesReceived.WithLabelValues("http").Add(float64(len(req.Body())))

	h.logger.Debugw("proxied http request",
		"method", method,
		"uri", string(ctx.RequestURI()),
		"status", status,
		"duration", duration,
	)
}

// handleConnect handles HTTPS CONNECT tunneling.
func (h *Handler) handleConnect(ctx *fasthttp.RequestCtx, start time.Time) {
	h.metrics.IncrementTunnels()
	defer h.metrics.DecrementTunnels()

	host := string(ctx.Host())

	// Ensure the host has a port
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "443")
	}

	// Connect to the destination
	destConn, err := h.dialHost(ctx, start, host)
	if err != nil {
		h.handleError(ctx, start, "CONNECT", "tunnel", err, "dial_failed")
		return
	}

	// Send 200 Connection Established
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.SetBodyRaw(nil)

	// Hijack the connection for bidirectional tunneling
	ctx.Hijack(func(clientConn net.Conn) {
		h.tunnel(clientConn, destConn, host, start)
	})
}

// dialHost connects to destination host.
func (h *Handler) dialHost(ctx *fasthttp.RequestCtx, start time.Time, host string) (destConn net.Conn, err error) {
	maxDials := h.config.MaxDialRetries + 1
	
	for i := 1; i <= maxDials; i++ {
		destConn, err = net.DialTimeout("tcp", host, h.config.DialTimeout)
		if err == nil {
			break
		}

		if i < maxDials {
			h.handleError(ctx, start, "CONNECT", "retry", err, fmt.Sprintf("#%d", i))

			time.Sleep(h.config.DialRetryDelay)
		}
	}

	return
}

// tunnel creates a bidirectional tunnel between client and destination.
func (h *Handler) tunnel(clientConn, destConn net.Conn, host string, start time.Time) {
	defer clientConn.Close()
	defer destConn.Close()

	var wg sync.WaitGroup
	var clientToServer, serverToClient int64

	// Client -> Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		clientToServer, _ = io.Copy(destConn, clientConn)
	}()

	// Server -> Client
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverToClient, _ = io.Copy(clientConn, destConn)
	}()

	wg.Wait()

	// Record metrics
	duration := time.Since(start).Seconds()
	h.metrics.RecordRequest("CONNECT", "200", "tunnel", duration)
	h.metrics.BytesSent.WithLabelValues("tunnel").Add(float64(serverToClient))
	h.metrics.BytesReceived.WithLabelValues("tunnel").Add(float64(clientToServer))

	h.logger.Debugw("tunnel closed",
		"host", host,
		"duration", duration,
		"bytes_sent", serverToClient,
		"bytes_received", clientToServer,
	)
}

// handleError handles and logs errors.
func (h *Handler) handleError(ctx *fasthttp.RequestCtx, start time.Time, method, reqType string, err error, reason string) {
	duration := time.Since(start).Seconds()

	ctx.Error(fmt.Sprintf("Proxy error: %v", err), fasthttp.StatusBadGateway)

	h.metrics.RecordRequest(method, "502", reqType, duration)
	h.metrics.RecordError(reqType, reason)

	h.logger.Warnw("proxy error",
		"method", method,
		"type", reqType,
		"error", err.Error(),
		"reason", reason,
	)
}

// removeHopByHopHeaders removes hop-by-hop headers from the header.
func removeHopByHopHeaders(header *fasthttp.RequestHeader) {
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
}

// removeResponseHopByHopHeaders removes hop-by-hop headers from response header.
func removeResponseHopByHopHeaders(header *fasthttp.ResponseHeader) {
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
}
