# Tham chiếu Types & Interfaces

Tài liệu này cung cấp thông tin toàn diện về tất cả các types, interfaces và cấu trúc dữ liệu được sử dụng trong BunnyHop.

## Core Types

### LoadBalanceStrategy

Định nghĩa chiến lược load balancing cho connection pools.

```go
type LoadBalanceStrategy int

const (
    RoundRobin         LoadBalanceStrategy = iota // Phân phối yêu cầu đều đặn
    Random                                       // Chọn node ngẫu nhiên
    LeastUsed                                   // Chọn node ít sử dụng nhất
    WeightedRoundRobin                          // Phân phối dựa trên trọng số
)
```

**Sử dụng:**
```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

### Logger

Interface cho các triển khai logging tùy chỉnh.

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

**Triển khai mặc định:**
```go
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
    log.Printf("[DEBUG] "+msg, args...)
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
    log.Printf("[INFO] "+msg, args...)
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
    log.Printf("[WARN] "+msg, args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
    log.Printf("[ERROR] "+msg, args...)
}
```

**Triển khai tùy chỉnh:**
```go
type CustomLogger struct {
    logger *zap.Logger
}

func (l *CustomLogger) Debug(msg string, args ...interface{}) {
    l.logger.Sugar().Debugf(msg, args...)
}

func (l *CustomLogger) Info(msg string, args ...interface{}) {
    l.logger.Sugar().Infof(msg, args...)
}

func (l *CustomLogger) Warn(msg string, args ...interface{}) {
    l.logger.Sugar().Warnf(msg, args...)
}

func (l *CustomLogger) Error(msg string, args ...interface{}) {
    l.logger.Sugar().Errorf(msg, args...)
}
```

## Configuration Types

### Config

Cấu trúc cấu hình client.

```go
type Config struct {
    URLs                []string      // Danh sách URL kết nối RabbitMQ
    ReconnectInterval   time.Duration // Thời gian giữa các lần thử kết nối lại
    MaxReconnectAttempt int           // Số lần thử kết nối lại tối đa
    DebugLog            bool          // Bật/tắt debug logging
    Logger              Logger        // Triển khai logger tùy chỉnh
    TLSConfig           *tls.Config   // Cấu hình TLS/SSL
}
```

**Chi tiết các trường:**
- `URLs`: Danh sách URL kết nối RabbitMQ (mặc định: `["amqp://localhost:5672"]`)
- `ReconnectInterval`: Thời gian giữa các lần thử kết nối lại (mặc định: `5s`)
- `MaxReconnectAttempt`: Số lần thử kết nối lại tối đa (mặc định: `10`)
- `DebugLog`: Bật/tắt debug logging (mặc định: `false`)
- `Logger`: Triển khai logger tùy chỉnh (mặc định: `DefaultLogger`)
- `TLSConfig`: Cấu hình TLS/SSL (tùy chọn)

**Ví dụ:**
```go
config := bunnyhop.Config{
    URLs:                []string{"amqp://user:pass@host:5672/"},
    ReconnectInterval:   3 * time.Second,
    MaxReconnectAttempt: 15,
    DebugLog:            true,
    Logger:              &CustomLogger{},
}
```

### PoolConfig

Cấu trúc cấu hình pool.

```go
type PoolConfig struct {
    URLs                []string           // Danh sách URL node RabbitMQ
    ReconnectInterval   time.Duration      // Thời gian giữa các lần thử kết nối lại
    MaxReconnectAttempt int                // Số lần thử kết nối lại tối đa
    HealthCheckInterval time.Duration      // Khoảng thời gian health check
    LoadBalanceStrategy LoadBalanceStrategy // Chiến lược load balancing
    DebugLog            bool               // Bật/tắt debug logging
    Logger              Logger             // Triển khai logger tùy chỉnh
    TLSConfig           *tls.Config        // Cấu hình TLS/SSL
}
```

**Chi tiết các trường:**
- `URLs`: Danh sách URL node RabbitMQ (bắt buộc)
- `ReconnectInterval`: Thời gian giữa các lần thử kết nối lại (mặc định: `5s`)
- `MaxReconnectAttempt`: Số lần thử kết nối lại tối đa (mặc định: `10`)
- `HealthCheckInterval`: Khoảng thời gian giữa các health check (mặc định: `30s`)
- `LoadBalanceStrategy`: Chiến lược load balancing (mặc định: `RoundRobin`)
- `DebugLog`: Bật/tắt debug logging (mặc định: `false`)
- `Logger`: Triển khai logger tùy chỉnh (mặc định: `DefaultLogger`)
- `TLSConfig`: Cấu hình TLS/SSL (tùy chọn)

**Ví dụ:**
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
    DebugLog:            false,
}
```

## Statistics Types

### PoolStats

Thống kê pool và thông tin sức khỏe.

```go
type PoolStats struct {
    TotalNodes          int                // Tổng số node
    HealthyNodes        int                // Số node khỏe mạnh
    UnhealthyNodes      int                // Số node không khỏe mạnh
    LastHealthCheck     time.Time          // Timestamp health check cuối cùng
    LoadBalanceStrategy LoadBalanceStrategy // Chiến lược load balancing hiện tại
}
```

**Sử dụng:**
```go
stats := pool.GetStats()
log.Printf("Tổng số node: %d", stats.TotalNodes)
log.Printf("Node khỏe mạnh: %d", stats.HealthyNodes)
log.Printf("Node không khỏe mạnh: %d", stats.UnhealthyNodes)
log.Printf("Health check cuối cùng: %v", stats.LastHealthCheck)
log.Printf("Chiến lược: %v", stats.LoadBalanceStrategy)
```

### ConnectionStats

Thống kê kết nối cá nhân.

```go
type ConnectionStats struct {
    URL           string    // URL kết nối
    IsConnected   bool      // Trạng thái kết nối
    LastUsed      time.Time // Timestamp sử dụng cuối cùng
    UsageCount    int64     // Tổng số lần sử dụng
    LastError     error     // Lỗi cuối cùng gặp phải
    ReconnectAttempts int   // Số lần thử kết nối lại
}
```

## Internal Types

### NodeConnection

Quản lý kết nối node nội bộ.

```go
type NodeConnection struct {
    URL           string         // URL node
    Client        *Client        // Instance client
    IsHealthy     bool           // Trạng thái sức khỏe
    LastUsed      time.Time      // Timestamp sử dụng cuối cùng
    UsageCount    int64          // Bộ đếm sử dụng
    Weight        int            // Trọng số node (cho các chiến lược có trọng số)
    mu            sync.RWMutex   // Mutex cho thread safety
}
```

**Methods:**
```go
func (nc *NodeConnection) MarkUsed()
func (nc *NodeConnection) IsAvailable() bool
func (nc *NodeConnection) GetHealth() bool
func (nc *NodeConnection) SetHealth(healthy bool)
```

### ConnectionState

Liệt kê trạng thái kết nối.

```go
type ConnectionState int

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateReconnecting
    StateFailed
)
```

## Error Types

### Custom Errors

BunnyHop định nghĩa một số custom error types để xử lý lỗi tốt hơn.

```go
var (
    ErrNoHealthyNodes    = errors.New("không có node khỏe mạnh nào")
    ErrPoolNotStarted    = errors.New("pool chưa được khởi động")
    ErrInvalidConfig     = errors.New("cấu hình không hợp lệ")
    ErrConnectionFailed  = errors.New("kết nối thất bại")
    ErrReconnectFailed   = errors.New("kết nối lại thất bại")
)
```

**Xử lý lỗi:**
```go
client, err := pool.GetClient()
if err != nil {
    switch {
    case errors.Is(err, bunnyhop.ErrNoHealthyNodes):
        log.Fatal("Không có node khỏe mạnh nào")
    case errors.Is(err, bunnyhop.ErrPoolNotStarted):
        log.Fatal("Pool chưa được khởi động")
    case errors.Is(err, bunnyhop.ErrInvalidConfig):
        log.Fatal("Cấu hình không hợp lệ")
    default:
        log.Printf("Lỗi không mong đợi: %v", err)
    }
    return
}
```

## Utility Types

### URLParser

Tiện ích để parse và validate RabbitMQ URLs.

```go
type URLParser struct {
    Scheme   string // amqp hoặc amqps
    Username string // Tên người dùng
    Password string // Mật khẩu
    Host     string // Tên máy chủ
    Port     int    // Số port
    VHost    string // Virtual host
}
```

**Sử dụng:**
```go
parser := &URLParser{}
if err := parser.Parse("amqp://user:pass@host:5672/vhost"); err != nil {
    log.Printf("URL không hợp lệ: %v", err)
    return
}

log.Printf("Host: %s, Port: %d, VHost: %s", parser.Host, parser.Port, parser.VHost)
```

### HealthChecker

Interface cho các triển khai health checking tùy chỉnh.

```go
type HealthChecker interface {
    CheckHealth(client *Client) bool
    GetHealthMetrics() HealthMetrics
}
```

**Triển khai mặc định:**
```go
type DefaultHealthChecker struct {
    timeout time.Duration
}

func (h *DefaultHealthChecker) CheckHealth(client *Client) bool {
    if !client.IsConnected() {
        return false
    }
    
    // Thực hiện health check cơ bản
    ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
    defer cancel()
    
    // Thử lấy channel
    if _, err := client.GetChannel(); err != nil {
        return false
    }
    
    return true
}
```

## Type Conversion Utilities

### String Conversion

```go
func (s LoadBalanceStrategy) String() string {
    switch s {
    case RoundRobin:
        return "RoundRobin"
    case Random:
        return "Random"
    case LeastUsed:
        return "LeastUsed"
    case WeightedRoundRobin:
        return "WeightedRoundRobin"
    default:
        return "Unknown"
    }
}
```

### JSON Marshaling

```go
func (s PoolStats) MarshalJSON() ([]byte, error) {
    type Alias PoolStats
    return json.Marshal(&struct {
        *Alias
        LastHealthCheck string `json:"last_health_check"`
    }{
        Alias:           (*Alias)(&s),
        LastHealthCheck: s.LastHealthCheck.Format(time.RFC3339),
    })
}
```

## Best Practices

### 1. Type Safety

Luôn sử dụng các types đã định nghĩa thay vì raw values:

```go
// Tốt
config := bunnyhop.Config{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}

// Tránh
config := bunnyhop.Config{
    LoadBalanceStrategy: 0, // Magic number
}
```

### 2. Interface Implementation

Triển khai interfaces một cách chính xác:

```go
type CustomLogger struct {
    logger *zap.Logger
}

// Đảm bảo tất cả methods được triển khai
func (l *CustomLogger) Debug(msg string, args ...interface{}) {
    l.logger.Sugar().Debugf(msg, args...)
}

func (l *CustomLogger) Info(msg string, args ...interface{}) {
    l.logger.Sugar().Infof(msg, args...)
}

func (l *CustomLogger) Warn(msg string, args ...interface{}) {
    l.logger.Sugar().Warnf(msg, args...)
}

func (l *CustomLogger) Error(msg string, args ...interface{}) {
    l.logger.Sugar().Errorf(msg, args...)
}
```

### 3. Error Handling

Sử dụng custom error types để xử lý lỗi tốt hơn:

```go
if err := pool.Start(); err != nil {
    if errors.Is(err, bunnyhop.ErrInvalidConfig) {
        log.Fatal("Lỗi cấu hình, kiểm tra cài đặt của bạn")
    }
    log.Fatalf("Khởi động pool thất bại: %v", err)
}
```

### 4. Configuration Validation

Validate cấu hình trước khi sử dụng:

```go
func validateConfig(config bunnyhop.Config) error {
    if len(config.URLs) == 0 {
        return errors.New("ít nhất phải có một URL")
    }
    
    if config.ReconnectInterval <= 0 {
        return errors.New("reconnect interval phải dương")
    }
    
    if config.MaxReconnectAttempt < 0 {
        return errors.New("max reconnect attempts không được âm")
    }
    
    return nil
}
```

## Ví dụ

### Triển khai Custom Logger

```go
package main

import (
    "log"
    "github.com/vanduc0209/bunnyhop"
)

type StructuredLogger struct {
    prefix string
}

func (l *StructuredLogger) Debug(msg string, args ...interface{}) {
    log.Printf("[DEBUG][%s] "+msg, append([]interface{}{l.prefix}, args...)...)
}

func (l *StructuredLogger) Info(msg string, args ...interface{}) {
    log.Printf("[INFO][%s] "+msg, append([]interface{}{l.prefix}, args...)...)
}

func (l *StructuredLogger) Warn(msg string, args ...interface{}) {
    log.Printf("[WARN][%s] "+msg, append([]interface{}{l.prefix}, args...)...)
}

func (l *StructuredLogger) Error(msg string, args ...interface{}) {
    log.Printf("[ERROR][%s] "+msg, append([]interface{}{l.prefix}, args...)...)
}

func main() {
    logger := &StructuredLogger{prefix: "BunnyHop"}
    
    config := bunnyhop.Config{
        URLs:   []string{"amqp://localhost:5672/"},
        Logger: logger,
    }
    
    client := bunnyhop.NewClient(config)
    // ... sử dụng client
}
```

### Custom Health Checker

```go
type CustomHealthChecker struct {
    checkInterval time.Duration
    lastCheck    time.Time
}

func (h *CustomHealthChecker) CheckHealth(client *Client) bool {
    now := time.Now()
    if now.Sub(h.lastCheck) < h.checkInterval {
        return true // Bỏ qua check nếu quá gần đây
    }
    
    h.lastCheck = now
    
    // Thực hiện health check tùy chỉnh
    if !client.IsConnected() {
        return false
    }
    
    // Các check tùy chỉnh bổ sung
    return true
}

func (h *CustomHealthChecker) GetHealthMetrics() HealthMetrics {
    return HealthMetrics{
        LastCheck: h.lastCheck,
        Status:    "healthy",
    }
}
```

## Bước tiếp theo

- Đọc [Tham chiếu API Client](client.md) cho các thao tác client
- Kiểm tra [Tham chiếu API Pool](pool.md) cho quản lý pool
- Xem lại [Hướng dẫn cấu hình](../configuration.md) cho các ví dụ sử dụng
- Khám phá [Ví dụ](../examples/) cho các ví dụ hoàn chỉnh có thể chạy được
