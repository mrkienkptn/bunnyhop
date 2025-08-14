# Tổng quan Kiến trúc

Tài liệu này cung cấp cái nhìn toàn diện về kiến trúc, nguyên tắc thiết kế và các thành phần nội bộ của BunnyHop.

## Tổng quan

BunnyHop là một thư viện Go được xây dựng trên `amqp091-go` cung cấp các abstraction cấp cao cho các thao tác RabbitMQ. Nó được thiết kế với các nguyên tắc sau:

- **Đơn giản**: API dễ sử dụng, ẩn đi sự phức tạp của RabbitMQ
- **Đáng tin cậy**: Quản lý kết nối, kết nối lại và failover tích hợp sẵn
- **Hiệu suất**: Connection pooling và load balancing hiệu quả
- **Mở rộng**: Logging, health checking và monitoring có thể plug-in

## Kiến trúc Cấp cao

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │   Application   │    │   Application   │
│     Layer       │    │     Layer       │    │     Layer       │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   BunnyHop      │    │   BunnyHop      │    │   BunnyHop      │
│    Client       │    │     Pool        │    │   Utilities     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   amqp091-go    │    │   amqp091-go    │    │   Standard      │
│   (RabbitMQ     │    │   (RabbitMQ     │    │   Library       │
│    Driver)      │    │    Driver)      │    │   Support       │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          ▼                      ▼
┌─────────────────┐    ┌─────────────────┐
│   RabbitMQ      │    │   RabbitMQ      │
│   Single Node   │    │   Cluster       │
└─────────────────┘    └─────────────────┘
```

## Các Thành phần Chính

### 1. Client

Type `Client` cung cấp một kết nối đơn đến RabbitMQ với khả năng tự động kết nối lại.

**Tính năng chính:**
- Quản lý kết nối đơn
- Tự động kết nối lại với logic retry có thể cấu hình
- Giám sát sức khỏe kết nối
- Các thao tác thread-safe

**Kiến trúc:**
```
┌─────────────────┐
│     Client      │
├─────────────────┤
│  Connection     │
│  Management     │
├─────────────────┤
│  Channel Pool   │
├─────────────────┤
│  Error Handler  │
├─────────────────┤
│  Logger         │
└─────────────────┘
```

**Cấu trúc nội bộ:**
```go
type Client struct {
    config     Config
    conn       *amqp.Connection
    channel    *amqp.Channel
    logger     Logger
    mu         sync.RWMutex
    isConnected bool
    // ... các trường khác
}
```

### 2. Pool

Type `Pool` quản lý nhiều kết nối RabbitMQ với khả năng load balancing và failover.

**Tính năng chính:**
- Hỗ trợ nhiều node
- Các chiến lược load balancing
- Failover tự động
- Giám sát sức khỏe
- Connection pooling

**Kiến trúc:**
```
┌─────────────────┐
│      Pool       │
├─────────────────┤
│  Node Manager   │
├─────────────────┤
│ Load Balancer   │
├─────────────────┤
│ Health Monitor  │
├─────────────────┤
│ Connection Pool │
├─────────────────┤
│  Error Handler  │
├─────────────────┤
│     Logger      │
└─────────────────┘
```

**Cấu trúc nội bộ:**
```go
type Pool struct {
    config     PoolConfig
    nodes      map[string]*NodeConnection
    strategy   LoadBalanceStrategy
    logger     Logger
    mu         sync.RWMutex
    // ... các trường khác
}
```

### 3. NodeConnection

Cấu trúc nội bộ để quản lý các kết nối node cá nhân trong một pool.

**Kiến trúc:**
```
┌─────────────────┐
│ NodeConnection  │
├─────────────────┤
│      URL        │
├─────────────────┤
│     Client      │
├─────────────────┤
│   Health Info   │
├─────────────────┤
│  Usage Stats    │
├─────────────────┤
│     Weight      │
└─────────────────┘
```

## Chiến lược Load Balancing

### 1. Round Robin (Mặc định)

Phân phối yêu cầu đều đặn qua tất cả các node khỏe mạnh theo thứ tự.

**Thuật toán:**
```go
func (p *Pool) getNextRoundRobin() *NodeConnection {
    p.currentIndex = (p.currentIndex + 1) % len(p.healthyNodes)
    return p.healthyNodes[p.currentIndex]
}
```

**Ưu điểm:**
- Phân phối có thể dự đoán
- Triển khai đơn giản
- Tốt cho hầu hết các trường hợp

### 2. Random

Chọn ngẫu nhiên một node khỏe mạnh cho mỗi yêu cầu.

**Thuật toán:**
```go
func (p *Pool) getNextRandom() *NodeConnection {
    if len(p.healthyNodes) == 0 {
        return nil
    }
    randomIndex := rand.Intn(len(p.healthyNodes))
    return p.healthyNodes[randomIndex]
}
```

**Ưu điểm:**
- Giảm thiểu các pattern có thể dự đoán
- Tốt cho các tình huống throughput cao
- Giảm thiểu quá tải cục bộ

### 3. Least Used

Chọn node có số lần sử dụng thấp nhất.

**Thuật toán:**
```go
func (p *Pool) getNextLeastUsed() *NodeConnection {
    var leastUsed *NodeConnection
    minUsage := int64(math.MaxInt64)
    
    for _, node := range p.healthyNodes {
        if node.UsageCount < minUsage {
            minUsage = node.UsageCount
            leastUsed = node
        }
    }
    
    return leastUsed
}
```

**Ưu điểm:**
- Tối ưu hóa việc sử dụng tài nguyên
- Load balancing tự động
- Tốt cho các thao tác tiêu tốn tài nguyên

### 4. Weighted Round Robin

Phân phối yêu cầu dựa trên trọng số của node.

**Thuật toán:**
```go
func (p *Pool) getNextWeightedRoundRobin() *NodeConnection {
    // Thuật toán phức tạp xem xét trọng số
    // và phân phối yêu cầu theo tỷ lệ
}
```

**Ưu điểm:**
- Linh hoạt và có thể tùy chỉnh
- Phù hợp cho các node có khả năng khác nhau
- Kiểm soát tốt việc phân phối

## Giám sát Sức khỏe

### Quy trình Health Check

Pool liên tục giám sát sức khỏe node thông qua một số cơ chế:

1. **Giám sát kết nối**: Theo dõi các kết nối bị đứt
2. **Health Check định kỳ**: Khoảng thời gian health check đều đặn
3. **Theo dõi lỗi**: Giám sát các thao tác thất bại
4. **Tự động kết nối lại**: Thử kết nối lại các node bị lỗi

**Luồng Health Check:**
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Start     │───▶│   Check     │───▶│   Update    │
│   Timer     │    │   Health    │    │   Status    │
└─────────────┘    └─────────────┘    └─────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │   Reconnect │
                   │   if Failed │
                   └─────────────┘
```

### Triển khai Health Check

```go
func (p *Pool) healthCheck() {
    ticker := time.NewTicker(p.config.HealthCheckInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            p.checkAllNodes()
        case <-p.stopChan:
            return
        }
    }
}

func (p *Pool) checkAllNodes() {
    for _, node := range p.nodes {
        go p.checkNodeHealth(node)
    }
}
```

## Quản lý Kết nối

### Vòng đời Kết nối

1. **Khởi tạo**: Tạo kết nối với cấu hình
2. **Thiết lập**: Kết nối đến server RabbitMQ
3. **Giám sát**: Theo dõi sức khỏe kết nối
4. **Kết nối lại**: Tự động kết nối lại khi thất bại
5. **Dọn dẹp**: Đóng kết nối một cách đúng đắn

**Trạng thái Kết nối:**
```
Disconnected ──▶ Connecting ──▶ Connected
     ▲                              │
     │                              ▼
     └── Reconnecting ◀─── Failed ◀──┘
```

### Logic Kết nối lại

```go
func (c *Client) reconnect() error {
    for attempt := 0; attempt < c.config.MaxReconnectAttempt; attempt++ {
        if err := c.connect(); err == nil {
            return nil
        }
        
        // Exponential backoff
        backoff := time.Duration(attempt+1) * c.config.ReconnectInterval
        time.Sleep(backoff)
    }
    
    return ErrReconnectFailed
}
```

## Xử lý Lỗi

### Các loại Lỗi

BunnyHop định nghĩa một số custom error types để xử lý lỗi tốt hơn:

```go
var (
    ErrNoHealthyNodes    = errors.New("không có node khỏe mạnh nào")
    ErrPoolNotStarted    = errors.New("pool chưa được khởi động")
    ErrInvalidConfig     = errors.New("cấu hình không hợp lệ")
    ErrConnectionFailed  = errors.New("kết nối thất bại")
    ErrReconnectFailed   = errors.New("kết nối lại thất bại")
)
```

### Chiến lược Xử lý Lỗi

1. **Graceful Degradation**: Tiếp tục hoạt động với các node có sẵn
2. **Retry Logic**: Tự động retry cho các lỗi tạm thời
3. **Circuit Breaker**: Ngăn chặn lỗi dây chuyền
4. **Error Propagation**: Thông báo lỗi rõ ràng để debug

## Thread Safety

### Mô hình Concurrency

BunnyHop sử dụng kết hợp mutexes và channels để đảm bảo thread safety:

```go
type Pool struct {
    mu    sync.RWMutex
    nodes map[string]*NodeConnection
    // ... các trường khác
}

func (p *Pool) GetClient() (*Client, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    // Các thao tác thread-safe
}
```

### Chiến lược Locking

- **Read-Write Mutexes**: Cho phép đọc đồng thời, ghi độc quyền
- **Fine-grained Locking**: Chỉ lock khi cần thiết
- **Deadlock Prevention**: Thứ tự lock nhất quán

## Cân nhắc Hiệu suất

### Connection Pooling

- **Tái sử dụng Kết nối**: Giảm thiểu overhead kết nối
- **Channel Pooling**: Quản lý channel hiệu quả
- **Lazy Initialization**: Tạo tài nguyên khi cần

### Quản lý Bộ nhớ

- **Tái sử dụng Object**: Giảm thiểu allocations
- **Buffer Pooling**: Tái sử dụng buffer cho messages
- **Garbage Collection**: Giảm thiểu áp lực GC

### Kỹ thuật Tối ưu hóa

1. **Tái sử dụng Kết nối**: Giữ kết nối sống
2. **Batch Operations**: Nhóm nhiều thao tác
3. **Async Processing**: Các thao tác không blocking
4. **Resource Limits**: Ngăn chặn cạn kiệt tài nguyên

## Giám sát và Khả năng Quan sát

### Thu thập Metrics

BunnyHop cung cấp metrics tích hợp sẵn để giám sát:

```go
type PoolStats struct {
    TotalNodes          int
    HealthyNodes        int
    UnhealthyNodes      int
    LastHealthCheck     time.Time
    LoadBalanceStrategy LoadBalanceStrategy
}
```

### Logging

Structured logging với các mức có thể cấu hình:

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

### Health Endpoints

Các endpoint health check tích hợp sẵn cho hệ thống giám sát:

```go
func (p *Pool) HealthCheck() HealthStatus {
    return HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Details:   p.GetStats(),
    }
}
```

## Bảo mật

### Xác thực

- **Username/Password**: Xác thực AMQP chuẩn
- **TLS/SSL**: Kết nối được mã hóa
- **Virtual Hosts**: Cô lập tài nguyên

### Bảo mật Mạng

- **Mã hóa Kết nối**: Hỗ trợ TLS
- **Kiểm soát Truy cập**: Quyền RabbitMQ
- **Cô lập Mạng**: Cân nhắc firewall

## Cân nhắc Triển khai

### Triển khai Single Node

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ BunnyHop    │
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  RabbitMQ   │
│   Single    │
└─────────────┘
```

### Triển khai Cluster

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ BunnyHop    │
│    Pool     │
└──────┬──────┘
       │
       ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  RabbitMQ   │    │  RabbitMQ   │    │  RabbitMQ   │
│   Node 1    │    │   Node 2    │    │   Node 3    │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Tích hợp Load Balancer

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ BunnyHop    │
│    Pool     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   HAProxy   │
└──────┬──────┘
       │
       ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  RabbitMQ   │    │  RabbitMQ   │    │  RabbitMQ   │
│   Node 1    │    │   Node 2    │    │   Node 3    │
└─────────────┘    └─────────────┘    └─────────────┘
```

## Cải tiến Tương lai

### Tính năng Dự kiến

1. **Circuit Breaker Pattern**: Xử lý lỗi nâng cao
2. **Rate Limiting**: Ngăn chặn quá tải RabbitMQ
3. **Metrics Export**: Tích hợp Prometheus/OpenTelemetry
4. **Distributed Tracing**: Theo dõi request qua các node
5. **Plugin System**: Kiến trúc có thể mở rộng

### Tiến hóa Kiến trúc

- **Hỗ trợ Microservices**: Pattern tích hợp tốt hơn
- **Event Sourcing**: Xử lý event nâng cao
- **Hỗ trợ CQRS**: Tách biệt Command/Query
- **Saga Pattern**: Hỗ trợ giao dịch phân tán

## Best Practices

### 1. Quản lý Kết nối

- Luôn đóng kết nối một cách đúng đắn
- Sử dụng connection pooling cho ứng dụng throughput cao
- Giám sát sức khỏe kết nối thường xuyên

### 2. Xử lý Lỗi

- Triển khai xử lý lỗi và retry logic đúng đắn
- Sử dụng custom error types để debug tốt hơn
- Log lỗi với context đầy đủ

### 3. Hiệu suất

- Chọn chiến lược load balancing phù hợp
- Giám sát metrics hiệu suất
- Tối ưu hóa dựa trên pattern workload

### 4. Bảo mật

- Sử dụng TLS cho triển khai production
- Triển khai xác thực đúng đắn
- Tuân theo best practices bảo mật

## Bước tiếp theo

- Đọc [Hướng dẫn Cấu hình](configuration.md) cho chi tiết setup
- Kiểm tra [Tham chiếu API](api/) cho tài liệu API chi tiết
- Xem lại [Ví dụ](examples/) cho các mẫu sử dụng
- Khám phá [Hướng dẫn Production](production.md) cho best practices triển khai
