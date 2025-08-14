# Tham chiếu API Pool

Type `Pool` cung cấp connection pool với nhiều node RabbitMQ, tự động failover và khả năng load balancing.

## Tổng quan

```go
type Pool struct {
    // Các trường private để sử dụng nội bộ
}
```

## Constructor

### NewPool

Tạo một connection pool mới với cấu hình được chỉ định.

```go
func NewPool(config PoolConfig) *Pool
```

**Tham số:**
- `config` - Cấu hình pool

**Trả về:**
- `*Pool` - Instance pool mới

**Ví dụ:**
```go
config := bunnyhop.PoolConfig{
    URLs: []string{
        "amqp://guest:guest@node1:5672/",
        "amqp://guest:guest@node2:5672/",
        "amqp://guest:guest@node3:5672/",
    },
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    HealthCheckInterval: 30 * time.Second,
    LoadBalanceStrategy: bunnyhop.RoundRobin,
    DebugLog:            true,
}

pool := bunnyhop.NewPool(config)
```

## Methods

### Start

Khởi động pool và thiết lập kết nối đến tất cả các node.

```go
func (p *Pool) Start() error
```

**Trả về:**
- `error` - Lỗi nếu khởi động pool thất bại

**Ví dụ:**
```go
if err := pool.Start(); err != nil {
    log.Fatalf("Khởi động pool thất bại: %v", err)
}
```

### GetClient

Lấy client sử dụng chiến lược load balancing đã cấu hình.

```go
func (p *Pool) GetClient() (*Client, error)
```

**Trả về:**
- `*Client` - Instance client RabbitMQ
- `error` - Lỗi nếu không có node khỏe mạnh nào

**Ví dụ:**
```go
client, err := pool.GetClient()
if err != nil {
    log.Printf("Lấy client thất bại: %v", err)
    return
}

// Sử dụng client cho các thao tác
queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
```

### GetStats

Lấy thống kê pool và thông tin sức khỏe.

```go
func (p *Pool) GetStats() PoolStats
```

**Trả về:**
- `PoolStats` - Cấu trúc thống kê pool

**Ví dụ:**
```go
stats := pool.GetStats()
log.Printf("Thống kê pool: %+v", stats)
log.Printf("Node khỏe mạnh: %d", stats.HealthyNodes)
log.Printf("Tổng số node: %d", stats.TotalNodes)
```

### Close

Đóng pool và tất cả các kết nối.

```go
func (p *Pool) Close() error
```

**Trả về:**
- `error` - Bất kỳ lỗi nào gặp phải trong quá trình dọn dẹp

**Ví dụ:**
```go
defer func() {
    if err := pool.Close(); err != nil {
        log.Printf("Lỗi đóng pool: %v", err)
    }
}()
```

### SetNodeWeight

Thiết lập trọng số cho một node cụ thể (cho chiến lược WeightedRoundRobin).

```go
func (p *Pool) SetNodeWeight(url string, weight int) error
```

**Tham số:**
- `url` - URL node để thiết lập trọng số
- `weight` - Giá trị trọng số (cao hơn = nhiều request hơn)

**Trả về:**
- `error` - Lỗi nếu thiết lập trọng số thất bại

**Ví dụ:**
```go
// Thiết lập trọng số khác nhau cho các node có khả năng khác nhau
pool.SetNodeWeight("amqp://node1:5672/", 3)  // Khả năng cao
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Khả năng trung bình
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Khả năng thấp hơn
```

### GetHealthyNodeCount

Lấy số lượng node khỏe mạnh hiện tại.

```go
func (p *Pool) GetHealthyNodeCount() int
```

**Trả về:**
- `int` - Số lượng node khỏe mạnh

**Ví dụ:**
```go
healthyCount := pool.GetHealthyNodeCount()
if healthyCount == 0 {
    log.Fatal("Không có node khỏe mạnh nào")
}
log.Printf("Node có sẵn: %d", healthyCount)
```

## Cấu hình

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

**Các trường:**
- `URLs` - Danh sách URL node RabbitMQ (bắt buộc)
- `ReconnectInterval` - Thời gian giữa các lần thử kết nối lại (mặc định: `5s`)
- `MaxReconnectAttempt` - Số lần thử kết nối lại tối đa (mặc định: `10`)
- `HealthCheckInterval` - Khoảng thời gian giữa các health check (mặc định: `30s`)
- `LoadBalanceStrategy` - Chiến lược load balancing (mặc định: `RoundRobin`)
- `DebugLog` - Bật/tắt debug logging (mặc định: `false`)
- `Logger` - Triển khai logger tùy chỉnh (mặc định: `DefaultLogger`)
- `TLSConfig` - Cấu hình TLS/SSL (tùy chọn)

### LoadBalanceStrategy

Các chiến lược load balancing có sẵn.

```go
const (
    RoundRobin         LoadBalanceStrategy = iota // Phân phối yêu cầu đều đặn
    Random                                       // Chọn node ngẫu nhiên
    LeastUsed                                   // Chọn node ít sử dụng nhất
    WeightedRoundRobin                          // Phân phối dựa trên trọng số
)
```

## Cấu trúc dữ liệu

### PoolStats

Cấu trúc thống kê pool.

```go
type PoolStats struct {
    TotalNodes     int           // Tổng số node
    HealthyNodes   int           // Số node khỏe mạnh
    UnhealthyNodes int           // Số node không khỏe mạnh
    LastHealthCheck time.Time    // Timestamp health check cuối cùng
    LoadBalanceStrategy LoadBalanceStrategy // Chiến lược hiện tại
}
```

### NodeConnection

Cấu trúc kết nối node nội bộ.

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

## Chiến lược Load Balancing

### Round Robin (Mặc định)

Phân phối yêu cầu đều đặn qua tất cả các node khỏe mạnh theo thứ tự.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

**Ưu điểm:**
- Phân phối đều đặn
- Đơn giản và có thể dự đoán
- Tốt cho hầu hết các trường hợp

### Random

Chọn ngẫu nhiên một node khỏe mạnh cho mỗi yêu cầu.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.Random,
}
```

**Ưu điểm:**
- Tốt cho các tình huống throughput cao
- Giảm thiểu các pattern có thể dự đoán
- Giảm quá tải cục bộ

### Least Used

Chọn node có số lần sử dụng thấp nhất.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.LeastUsed,
}
```

**Ưu điểm:**
- Tối ưu hóa việc sử dụng tài nguyên
- Load balancing tự động
- Tốt cho các thao tác tiêu tốn tài nguyên

### Weighted Round Robin

Phân phối yêu cầu dựa trên trọng số của node.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.WeightedRoundRobin,
}

// Thiết lập trọng số cho các node có khả năng khác nhau
pool.SetNodeWeight("amqp://node1:5672/", 3)  // Khả năng cao
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Khả năng trung bình
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Khả năng thấp hơn
```

**Ưu điểm:**
- Linh hoạt và có thể tùy chỉnh
- Phù hợp cho các node có khả năng khác nhau
- Kiểm soát tốt việc phân phối

## Giám sát sức khỏe

### Health Check tự động

Pool tự động giám sát sức khỏe node:

- **Giám sát kết nối**: Theo dõi các kết nối bị đứt
- **Health Check**: Health check định kỳ trên tất cả các node
- **Tự động kết nối lại**: Tự động kết nối lại đến các node bị lỗi
- **Load Balancing**: Chỉ định tuyến yêu cầu đến các node khỏe mạnh

### Cấu hình Health Check

```go
config := bunnyhop.PoolConfig{
    HealthCheckInterval: 15 * time.Second, // Health check thường xuyên hơn
    ReconnectInterval:   3 * time.Second,  // Kết nối lại nhanh hơn
    MaxReconnectAttempt: 15,               // Nhiều lần thử hơn
}
```

## Xử lý lỗi

### Lỗi kết nối

Pool tự động xử lý:

- **Lỗi mạng**: Tự động kết nối lại
- **Lỗi node**: Failover đến các node khỏe mạnh
- **Lỗi xác thực**: Retry với exponential backoff
- **Lỗi TLS**: Fallback về non-TLS nếu được cấu hình

### Các tình huống lỗi

```go
client, err := pool.GetClient()
if err != nil {
    switch {
    case errors.Is(err, ErrNoHealthyNodes):
        log.Fatal("Không có node khỏe mạnh nào")
    case errors.Is(err, ErrPoolNotStarted):
        log.Fatal("Pool chưa được khởi động")
    default:
        log.Printf("Lỗi không mong đợi: %v", err)
    }
    return
}
```

## Best Practices

1. **Luôn đóng pool** sử dụng `defer pool.Close()`
2. **Kiểm tra sức khỏe pool** trước khi thực hiện các thao tác sử dụng `GetHealthyNodeCount()`
3. **Monitor thống kê** sử dụng `GetStats()` cho production monitoring
4. **Sử dụng chiến lược phù hợp** dựa trên yêu cầu workload của bạn
5. **Thiết lập trọng số node** cho WeightedRoundRobin để tối ưu hóa phân phối
6. **Xử lý lỗi một cách nhẹ nhàng** và triển khai logic retry cho các thao tác quan trọng

## Ví dụ

### Sử dụng Pool cơ bản

```go
package main

import (
    "log"
    "time"
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    config := bunnyhop.PoolConfig{
        URLs: []string{
            "amqp://guest:guest@node1:5672/",
            "amqp://guest:guest@node2:5672/",
            "amqp://guest:guest@node3:5672/",
        },
        HealthCheckInterval: 30 * time.Second,
        LoadBalanceStrategy: bunnyhop.RoundRobin,
        DebugLog:            true,
    }
    
    pool := bunnyhop.NewPool(config)
    defer pool.Close()
    
    if err := pool.Start(); err != nil {
        log.Fatalf("Khởi động pool thất bại: %v", err)
    }
    
    // Chờ pool sẵn sàng
    time.Sleep(2 * time.Second)
    
    // Lấy client sử dụng load balancing
    client, err := pool.GetClient()
    if err != nil {
        log.Fatalf("Lấy client thất bại: %v", err)
    }
    
    // Sử dụng client
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Printf("Khai báo queue thất bại: %v", err)
        return
    }
    
    log.Printf("Queue đã được khai báo: %s", queue.Name)
    
    // Lấy thống kê pool
    stats := pool.GetStats()
    log.Printf("Thống kê pool: %+v", stats)
}
```

### Cấu hình Pool cho Production

```go
func createProductionPool() *bunnyhop.Pool {
    config := bunnyhop.PoolConfig{
        URLs: []string{
            "amqp://prod_user:prod_pass@prod1:5672/",
            "amqp://prod_user:prod_pass@prod2:5672/",
            "amqp://prod_user:prod_pass@prod3:5672/",
        },
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        HealthCheckInterval: 15 * time.Second,
        LoadBalanceStrategy: bunnyhop.WeightedRoundRobin,
        DebugLog:            false,
    }
    
    pool := bunnyhop.NewPool(config)
    
    // Thiết lập trọng số cho các node có khả năng khác nhau
    pool.SetNodeWeight("amqp://prod1:5672/", 3)  // Khả năng cao
    pool.SetNodeWeight("amqp://prod2:5672/", 2)  // Khả năng trung bình
    pool.SetNodeWeight("amqp://prod3:5672/", 1)  // Khả năng thấp hơn
    
    return pool
}
```

## Bước tiếp theo

- Đọc [Tham chiếu API Client](client.md) cho các thao tác client
- Kiểm tra [Hướng dẫn cấu hình](../configuration.md) cho các tùy chọn chi tiết
- Xem lại [Triển khai Production](../production.md) cho best practices
- Khám phá [Ví dụ](../examples/) cho các mẫu sử dụng
