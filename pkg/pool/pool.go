// Package pool provides connection pooling for upstream HTTP requests.
package pool

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/yigitkonur/proxy-http-forward/pkg/config"
)

// Pool manages a pool of fasthttp clients for making upstream requests.
type Pool struct {
	pool   sync.Pool
	config config.ProxyConfig
}

// New creates a new connection pool with the given configuration.
func New(cfg config.ProxyConfig) *Pool {
	p := &Pool{
		config: cfg,
	}

	p.pool = sync.Pool{
		New: func() interface{} {
			return p.newClient()
		},
	}

	return p
}

// newClient creates a new fasthttp client with the pool configuration.
func (p *Pool) newClient() *fasthttp.Client {
	return &fasthttp.Client{
		// Connection settings
		MaxConnsPerHost:     p.config.MaxIdleConns,
		MaxIdleConnDuration: time.Minute * 5,

		// Timeout settings
		ReadTimeout:  p.config.ResponseTimeout,
		WriteTimeout: p.config.ResponseTimeout,

		// Dialer settings
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,

		// Disable automatic redirect following (proxy should forward as-is)
		NoDefaultUserAgentHeader: true,

		// Allow connections to any host
		DisablePathNormalizing: true,
	}
}

// Get retrieves a client from the pool.
func (p *Pool) Get() *fasthttp.Client {
	return p.pool.Get().(*fasthttp.Client)
}

// Put returns a client to the pool.
func (p *Pool) Put(client *fasthttp.Client) {
	p.pool.Put(client)
}

// Do executes an HTTP request using a pooled client.
func (p *Pool) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	client := p.Get()
	defer p.Put(client)
	return client.Do(req, resp)
}

// DoTimeout executes an HTTP request with a timeout using a pooled client.
func (p *Pool) DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	client := p.Get()
	defer p.Put(client)
	return client.DoTimeout(req, resp, timeout)
}
