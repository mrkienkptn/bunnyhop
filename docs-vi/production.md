# Triển khai Production

Hướng dẫn này bao gồm các best practices để triển khai BunnyHop trong môi trường production.

## Tổng quan

Triển khai production yêu cầu cân nhắc cẩn thận về:
- **High Availability** - Đảm bảo tính liên tục của dịch vụ
- **Performance** - Tối ưu hóa throughput và latency
- **Monitoring** - Theo dõi sức khỏe hệ thống và metrics
- **Security** - Bảo vệ chống truy cập trái phép
- **Scalability** - Xử lý tải tăng một cách nhẹ nhàng

## Khuyến nghị kiến trúc

### Kích thước Connection Pool

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
    DebugLog:            false, // Tắt trong production
}
```

**Hướng dẫn kích thước Pool:**
- **Ứng dụng nhỏ**: 2-3 kết nối mỗi pool
- **Ứng dụng trung bình**: 5-10 kết nối mỗi pool
- **Ứng dụng lớn**: 10-20 kết nối mỗi pool
- **Microservices**: 1-2 kết nối mỗi service

### Chiến lược Load Balancing

Chọn dựa trên workload của bạn:

- **RoundRobin**: Phân phối đều, tốt cho hầu hết các trường hợp
- **Random**: Tốt cho các tình huống throughput cao
- **LeastUsed**: Tốt nhất cho các thao tác tiêu tốn tài nguyên
- **WeightedRoundRobin**: Phân phối tùy chỉnh dựa trên khả năng node

```go
// Ví dụ: Phân phối có trọng số cho các node có khả năng khác nhau
pool.SetNodeWeight("amqp://node1:5672/", 3)  // Khả năng cao
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Khả năng trung bình
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Khả năng thấp hơn
```

## Best Practices cấu hình

### Biến môi trường

```bash
# Cấu hình production
export RABBITMQ_URLS="amqp://user:pass@node1:5672/,amqp://user:pass@node2:5672/,amqp://user:pass@node3:5672/"
export RABBITMQ_DEBUG="false"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
export RABBITMQ_HEALTH_CHECK_INTERVAL="30s"
export RABBITMQ_LOAD_BALANCE_STRATEGY="WeightedRoundRobin"
```

### File cấu hình

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

## Cân nhắc bảo mật

### Cấu hình TLS/SSL

```go
// Bật TLS cho kết nối bảo mật
config := bunnyhop.Config{
    URLs: []string{"amqps://user:pass@node1:5671/"},
    TLSConfig: &tls.Config{
        RootCAs:            rootCAs,
        Certificates:       []tls.Certificate{cert},
        InsecureSkipVerify: false,
    },
}
```

### Xác thực

```go
// Sử dụng xác thực mạnh
config := bunnyhop.Config{
    URLs: []string{
        "amqp://service_user:strong_password@node1:5672/",
        "amqp://service_user:strong_password@node2:5672/",
        "amqp://service_user:strong_password@node3:5672/",
    },
}
```

### Bảo mật mạng

- **Quy tắc Firewall**: Hạn chế truy cập vào các port RabbitMQ (5672, 5671)
- **VPC/Network Isolation**: Sử dụng mạng riêng cho giao tiếp nội bộ
- **VPN Access**: Yêu cầu VPN để truy cập bên ngoài vào các interface quản lý

## Monitoring và Health Checks

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

### Thu thập Metrics

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
            Help: "Tổng số kết nối RabbitMQ",
        },
        []string{"status"},
    )
    
    messagePublished = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "rabbitmq_messages_published_total",
            Help: "Tổng số tin nhắn đã gửi",
        },
        []string{"exchange", "routing_key"},
    )
)

func init() {
    prometheus.MustRegister(connectionTotal)
    prometheus.MustRegister(messagePublished)
}
```

// Cập nhật metrics trong ứng dụng của bạn
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
// Structured logging cho production
import (
    "go.uber.org/zap"
)

func setupLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{"stdout", "/var/log/app/rabbitmq.log"}
    config.ErrorOutputPaths = []string{"stderr", "/var/log/app/rabbitmq-error.log"}
    
    logger, err := config.Build()
    if err != nil {
        log.Fatalf("Tạo logger thất bại: %v", err)
    }
    
    return logger
}

// Triển khai logger tùy chỉnh
type ProductionLogger struct {
    logger *zap.Logger
}

func (l *ProductionLogger) Info(msg string, args ...interface{}) {
    l.logger.Sugar().Infof(msg, args...)
}

func (l *ProductionLogger) Error(msg string, args ...interface{}) {
    l.logger.Sugar().Errorf(msg, args...)
}

// Sử dụng trong cấu hình
config := bunnyhop.Config{
    URLs:   []string{"amqp://user:pass@node1:5672/"},
    Logger: &ProductionLogger{logger: setupLogger()},
}
```

## Tối ưu hóa Performance

### Connection Pooling

```go
// Tối ưu hóa pool cho throughput cao
config := bunnyhop.PoolConfig{
    URLs:                urls,
    HealthCheckInterval: 15 * time.Second, // Health check thường xuyên hơn
    LoadBalanceStrategy: bunnyhop.LeastUsed, // Tốt hơn cho tải cao
}

// Pre-warm connections
pool := bunnyhop.NewPool(config)
if err := pool.Start(); err != nil {
    log.Fatalf("Khởi động pool thất bại: %v", err)
}

// Chờ tất cả kết nối được thiết lập
time.Sleep(5 * time.Second)
```

### Message Batching

```go
// Batch messages để có performance tốt hơn
func publishBatch(client *bunnyhop.Client, exchange string, messages []string) error {
    channel, err := client.GetChannel()
    if err != nil {
        return err
    }
    
    // Bật publisher confirms
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
    
    // Chờ tất cả confirms
    if confirmed := <-confirms; !confirmed.Ack {
        return fmt.Errorf("xác nhận batch tin nhắn thất bại")
    }
    
    return nil
}
```

## Chiến lược triển khai

### Blue-Green Deployment

```go
// Graceful shutdown cho zero-downtime deployments
func gracefulShutdown(pool *bunnyhop.Pool) {
    // Dừng nhận request mới
    log.Println("Dừng kết nối mới...")
    
    // Chờ các thao tác hiện tại hoàn thành
    time.Sleep(10 * time.Second)
    
    // Đóng pool
    if err := pool.Close(); err != nil {
        log.Printf("Lỗi đóng pool: %v", err)
    }
    
    log.Println("Pool đã đóng một cách nhẹ nhàng")
}

// Xử lý tín hiệu shutdown
func main() {
    pool := setupPool()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        gracefulShutdown(pool)
        os.Exit(0)
    }()
    
    // Logic ứng dụng của bạn ở đây
}
```

### Container Deployment

```dockerfile
# Dockerfile cho production
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

## Monitoring và Alerting

### Các Metrics chính cần theo dõi

- **Connection Count**: Tổng số kết nối đang hoạt động
- **Healthy Nodes**: Số lượng RabbitMQ node có sẵn
- **Message Throughput**: Tin nhắn mỗi giây
- **Error Rate**: Phần trăm thao tác thất bại
- **Response Time**: Latency của kết nối và thao tác

### Quy tắc Alerting

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
          summary: "Không có RabbitMQ node khỏe mạnh nào"
          
      - alert: HighErrorRate
        expr: rate(rabbitmq_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Phát hiện tỷ lệ lỗi cao"
          
      - alert: ConnectionPoolExhausted
        expr: rabbitmq_connection_pool_usage > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Sử dụng connection pool cao"
```

## Xử lý sự cố

### Các vấn đề Production thường gặp

1. **Connection Exhaustion**
   - Tăng kích thước pool
   - Triển khai connection timeouts
   - Monitor việc sử dụng connection

2. **High Latency**
   - Kiểm tra kết nối mạng
   - Tối ưu hóa kích thước tin nhắn
   - Sử dụng message batching

3. **Memory Leaks**
   - Monitor số lượng goroutine
   - Kiểm tra các kết nối chưa đóng
   - Triển khai cleanup đúng cách

### Debug Mode

```go
// Bật debug mode tạm thời để xử lý sự cố
config := bunnyhop.Config{
    URLs:     urls,
    DebugLog: true, // Chỉ bật khi cần thiết
}
```

## Bước tiếp theo

- Xem lại [Hướng dẫn Monitoring](monitoring.md) để biết chi tiết về metrics
- Kiểm tra [Hướng dẫn xử lý sự cố](troubleshooting.md) cho các vấn đề thường gặp
- Khám phá [Ví dụ nâng cao](../examples/advanced.md) cho các tình huống phức tạp
- Đọc [Tham chiếu API](../api/) để biết tài liệu hoàn chỉnh 