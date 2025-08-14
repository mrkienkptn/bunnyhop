# Production Deployment

This guide covers best practices for deploying BunnyHop in production environments.

## Overview

Production deployments require careful consideration of:
- **High Availability** - Ensuring service continuity
- **Performance** - Optimizing for throughput and latency
- **Monitoring** - Tracking system health and metrics
- **Security** - Protecting against unauthorized access
- **Scalability** - Handling increased load gracefully

## Architecture Recommendations

### Connection Pool Sizing

```go
config := bunnyhop.PoolConfig{
    URLs: []string{
        "amqp://user:pass@node1:5672/",
        "amqp://user:pass@node2:5672/",
        "amqp://user:pass@node3:5672/",
    },
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    HealthCheckInterval: 30 * time.Second,
    LoadBalanceStrategy: bunnyhop.WeightedRoundRobin,
    DebugLog:            false, // Disable in production
}
```

**Pool Size Guidelines:**
- **Small applications**: 2-3 connections per pool
- **Medium applications**: 5-10 connections per pool
- **Large applications**: 10-20 connections per pool
- **Microservices**: 1-2 connections per service

### Load Balancing Strategy

Choose based on your workload:

- **RoundRobin**: Even distribution, good for most cases
- **Random**: Good for high-throughput scenarios
- **LeastUsed**: Best for resource-intensive operations
- **WeightedRoundRobin**: Custom distribution based on node capacity

```go
// Example: Weighted distribution for different node capacities
pool.SetNodeWeight("amqp://node1:5672/", 3)  // High capacity
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Medium capacity
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Lower capacity
```

## Configuration Best Practices

### Environment Variables

```bash
# Production configuration
export RABBITMQ_URLS="amqp://user:pass@node1:5672/,amqp://user:pass@node2:5672/,amqp://user:pass@node3:5672/"
export RABBITMQ_DEBUG="false"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
export RABBITMQ_HEALTH_CHECK_INTERVAL="30s"
export RABBITMQ_LOAD_BALANCE_STRATEGY="WeightedRoundRobin"
```

### Configuration File

```yaml
# config/production.yaml
rabbitmq:
  urls:
    - "amqp://user:pass@node1:5672/"
    - "amqp://user:pass@node2:5672/"
    - "amqp://user:pass@node3:5672/"
  
  pool:
    reconnect_interval: "5s"
    max_reconnect_attempts: 10
    health_check_interval: "30s"
    load_balance_strategy: "WeightedRoundRobin"
    debug_log: false
  
  security:
    tls_enabled: true
    ca_cert_path: "/etc/ssl/certs/ca-certificates.crt"
    client_cert_path: "/etc/ssl/certs/client.crt"
    client_key_path: "/etc/ssl/certs/client.key"
  
  monitoring:
    metrics_enabled: true
    health_check_endpoint: "/health"
    prometheus_endpoint: "/metrics"
```

## Security Considerations

### TLS/SSL Configuration

```go
// Enable TLS for secure connections
config := bunnyhop.Config{
    URLs: []string{"amqps://user:pass@node1:5671/"},
    TLSConfig: &tls.Config{
        RootCAs:            rootCAs,
        Certificates:       []tls.Certificate{cert},
        InsecureSkipVerify: false,
    },
}
```

### Authentication

```go
// Use strong authentication
config := bunnyhop.Config{
    URLs: []string{
        "amqp://service_user:strong_password@node1:5672/",
        "amqp://service_user:strong_password@node2:5672/",
        "amqp://service_user:strong_password@node3:5672/",
    },
}
```

### Network Security

- **Firewall Rules**: Restrict access to RabbitMQ ports (5672, 5671)
- **VPC/Network Isolation**: Use private networks for internal communication
- **VPN Access**: Require VPN for external access to management interfaces

## Monitoring and Health Checks

### Health Check Endpoint

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/vanduc0209/bunnyhop"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    stats := pool.GetStats()
    
    health := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().UTC(),
        "pool_stats": stats,
        "healthy_nodes": pool.GetHealthyNodeCount(),
    }
    
    if stats.HealthyNodes == 0 {
        health["status"] = "unhealthy"
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(health)
}

func main() {
    http.HandleFunc("/health", healthHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Metrics Collection

```go
// Prometheus metrics
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    connectionTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rabbitmq_connections_total",
            Help: "Total number of RabbitMQ connections",
        },
        []string{"status"},
    )
    
    messagePublished = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rabbitmq_messages_published_total",
            Help: "Total number of messages published",
        },
        []string{"exchange", "routing_key"},
    )
)

func init() {
    prometheus.MustRegister(connectionTotal)
    prometheus.MustRegister(messagePublished)
}

// Update metrics in your application
func publishMessage(client *bunnyhop.Client, exchange, routingKey string, message []byte) error {
    err := client.PublishMessage(exchange, routingKey, false, false, amqp.Publishing{
        Body: message,
    })
    
    if err == nil {
        messagePublished.WithLabelValues(exchange, routingKey).Inc()
    }
    
    return err
}
```

### Logging

```go
// Structured logging for production
import (
    "go.uber.org/zap"
)

func setupLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"stdout", "/var/log/app/rabbitmq.log"}
    config.ErrorOutputPaths = []string{"stderr", "/var/log/app/rabbitmq-error.log"}
    
    logger, err := config.Build()
    if err != nil {
        log.Fatalf("Failed to create logger: %v", err)
    }
    
    return logger
}

// Custom logger implementation
type ProductionLogger struct {
    logger *zap.Logger
}

func (l *ProductionLogger) Info(msg string, args ...interface{}) {
    l.logger.Sugar().Infof(msg, args...)
}

func (l *ProductionLogger) Error(msg string, args ...interface{}) {
    l.logger.Sugar().Errorf(msg, args...)
}

// Use in configuration
config := bunnyhop.Config{
    URLs:   []string{"amqp://user:pass@node1:5672/"},
    Logger: &ProductionLogger{logger: setupLogger()},
}
```

## Performance Optimization

### Connection Pooling

```go
// Optimize pool for high throughput
config := bunnyhop.PoolConfig{
    URLs:                urls,
    HealthCheckInterval: 15 * time.Second, // More frequent health checks
    LoadBalanceStrategy: bunnyhop.LeastUsed, // Better for high load
}

// Pre-warm connections
pool := bunnyhop.NewPool(config)
if err := pool.Start(); err != nil {
    log.Fatalf("Failed to start pool: %v", err)
}

// Wait for all connections to be established
time.Sleep(5 * time.Second)
```

### Message Batching

```go
// Batch messages for better performance
func publishBatch(client *bunnyhop.Client, exchange string, messages []string) error {
    channel, err := client.GetChannel()
    if err != nil {
        return err
    }
    
    // Enable publisher confirms
    err = channel.Confirm(false)
    if err != nil {
        return err
    }
    
    confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
    
    for _, msg := range messages {
        err := channel.Publish(exchange, "", false, false, amqp.Publishing{
            Body: []byte(msg),
        })
        if err != nil {
            return err
        }
    }
    
    // Wait for all confirms
    if confirmed := <-confirms; !confirmed.Ack {
        return fmt.Errorf("failed to confirm message batch")
    }
    
    return nil
}
```

## Deployment Strategies

### Blue-Green Deployment

```go
// Graceful shutdown for zero-downtime deployments
func gracefulShutdown(pool *bunnyhop.Pool) {
    // Stop accepting new requests
    log.Println("Stopping new connections...")
    
    // Wait for existing operations to complete
    time.Sleep(10 * time.Second)
    
    // Close the pool
    if err := pool.Close(); err != nil {
        log.Printf("Error closing pool: %v", err)
    }
    
    log.Println("Pool closed gracefully")
}

// Handle shutdown signals
func main() {
    pool := setupPool()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        gracefulShutdown(pool)
        os.Exit(0)
    }()
    
    // Your application logic here
}
```

### Container Deployment

```dockerfile
# Dockerfile for production
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/main .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./main"]
```

```yaml
# docker-compose.production.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - RABBITMQ_URLS=${RABBITMQ_URLS}
      - RABBITMQ_DEBUG=false
    depends_on:
      - rabbitmq
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Monitoring and Alerting

### Key Metrics to Monitor

- **Connection Count**: Total active connections
- **Healthy Nodes**: Number of available RabbitMQ nodes
- **Message Throughput**: Messages per second
- **Error Rate**: Failed operations percentage
- **Response Time**: Connection and operation latency

### Alerting Rules

```yaml
# prometheus/alerting-rules.yml
groups:
  - name: rabbitmq
    rules:
      - alert: NoHealthyNodes
        expr: rabbitmq_healthy_nodes == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "No healthy RabbitMQ nodes available"
          
      - alert: HighErrorRate
        expr: rate(rabbitmq_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          
      - alert: ConnectionPoolExhausted
        expr: rabbitmq_connection_pool_usage > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Connection pool usage is high"
```

## Troubleshooting

### Common Production Issues

1. **Connection Exhaustion**
   - Increase pool size
   - Implement connection timeouts
   - Monitor connection usage

2. **High Latency**
   - Check network connectivity
   - Optimize message size
   - Use message batching

3. **Memory Leaks**
   - Monitor goroutine count
   - Check for unclosed connections
   - Implement proper cleanup

### Debug Mode

```go
// Enable debug mode temporarily for troubleshooting
config := bunnyhop.Config{
    URLs:     urls,
    DebugLog: true, // Enable only when needed
}
```

## Next Steps

- Review [Monitoring Guide](monitoring.md) for detailed metrics
- Check [Troubleshooting Guide](troubleshooting.md) for common issues
- Explore [Advanced Examples](../examples/advanced.md) for complex scenarios
- Read [API Reference](../api/) for complete documentation 