# proxy-http-forward - Dockerfile
# Multi-stage build for minimal image size

# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.buildDate=${BUILD_DATE}" \
    -o /proxy \
    ./cmd/proxy

# ============================================
# Stage 2: Runtime
# ============================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' proxy

# Copy binary from builder
COPY --from=builder /proxy /usr/local/bin/proxy

# Copy default config
COPY config.yaml /etc/proxy/config.yaml

# Set ownership
RUN chown -R proxy:proxy /etc/proxy

# Switch to non-root user
USER proxy

# Expose ports
# 8080: Proxy server
# 9090: Prometheus metrics
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/metrics || exit 1

# Default config location
ENV PROXY_SERVER_ADDRESS=":8080"
ENV PROXY_METRICS_ADDRESS=":9090"

# Run the proxy
ENTRYPOINT ["proxy"]
CMD ["-config", "/etc/proxy/config.yaml"]
