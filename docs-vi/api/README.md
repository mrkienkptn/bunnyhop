# Tham chiếu API

Phần này cung cấp tài liệu API toàn diện cho BunnyHop.

## API Có sẵn

### Core APIs

- **[Client API](client.md)** - Tham chiếu đầy đủ cho type `Client`
- **[Pool API](pool.md)** - Tham chiếu đầy đủ cho type `Pool`  
- **[Types & Interfaces](types.md)** - Tất cả types, interfaces và cấu trúc dữ liệu

### Tham khảo Nhanh

| Thành phần | Mục đích | Methods chính |
|------------|----------|---------------|
| **Client** | Kết nối RabbitMQ đơn | `Connect()`, `DeclareQueue()`, `PublishMessage()`, `Consume()` |
| **Pool** | Nhiều kết nối với load balancing | `Start()`, `GetClient()`, `GetStats()`, `Close()` |

### Cấu hình

| Type | Mô tả | Các trường chính |
|------|-------|------------------|
| **Config** | Cấu hình client | `URLs`, `ReconnectInterval`, `MaxReconnectAttempt`, `DebugLog` |
| **PoolConfig** | Cấu hình pool | `URLs`, `LoadBalanceStrategy`, `HealthCheckInterval`, `Logger` |

### Chiến lược Load Balancing

| Chiến lược | Mô tả | Trường hợp sử dụng |
|------------|-------|-------------------|
| **RoundRobin** | Phân phối đều đặn theo thứ tự | Mục đích chung, có thể dự đoán |
| **Random** | Chọn ngẫu nhiên | Throughput cao, phân phối tải |
| **LeastUsed** | Chọn node ít sử dụng nhất | Tối ưu hóa tài nguyên |
| **WeightedRoundRobin** | Phân phối dựa trên trọng số | Khả năng node khác nhau |

## Bắt đầu

1. **Chọn cách tiếp cận**:
   - Sử dụng `Client` cho các tình huống đơn giản, kết nối đơn
   - Sử dụng `Pool` cho các tình huống high-availability, multi-node

2. **Cấu hình setup**:
   - Thiết lập URL kết nối
   - Chọn chiến lược load balancing (cho pools)
   - Cấu hình cài đặt kết nối lại

3. **Triển khai xử lý lỗi**:
   - Xử lý lỗi kết nối
   - Triển khai logic retry
   - Giám sát trạng thái sức khỏe

## Ví dụ

### Sử dụng Client cơ bản

```go
config := bunnyhop.Config{
    URLs: []string{"amqp://localhost:5672/"},
    DebugLog: true,
}

client := bunnyhop.NewClient(config)
defer client.Close()

if err := client.Connect(context.Background()); err != nil {
    log.Fatal(err)
}
```

### Pool với Load Balancing

```go
config := bunnyhop.PoolConfig{
    URLs: []string{
        "amqp://node1:5672/",
        "amqp://node2:5672/",
        "amqp://node3:5672/",
    },
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}

pool := bunnyhop.NewPool(config)
defer pool.Close()

if err := pool.Start(); err != nil {
    log.Fatal(err)
}

client, err := pool.GetClient()
if err != nil {
    log.Fatal(err)
}
```

## Xử lý Lỗi

BunnyHop cung cấp custom error types để xử lý lỗi tốt hơn:

```go
var (
    ErrNoHealthyNodes    = errors.New("không có node khỏe mạnh nào")
    ErrPoolNotStarted    = errors.New("pool chưa được khởi động")
    ErrInvalidConfig     = errors.New("cấu hình không hợp lệ")
    ErrConnectionFailed  = errors.New("kết nối thất bại")
    ErrReconnectFailed   = errors.New("kết nối lại thất bại")
)
```

## Thread Safety

- **Client**: Thread-safe cho các thao tác đồng thời
- **Pool**: Thread-safe với read-write mutexes
- **Tất cả thao tác**: An toàn cho truy cập đồng thời

## Cân nhắc Hiệu suất

- **Tái sử dụng kết nối**: Giảm thiểu overhead kết nối
- **Channel pooling**: Quản lý channel hiệu quả
- **Load balancing**: Phân phối tải qua các node
- **Giám sát sức khỏe**: Failover tự động

## Bước tiếp theo

- Đọc tài liệu API cá nhân để có thông tin chi tiết
- Kiểm tra [Hướng dẫn Cấu hình](../configuration.md) cho chi tiết setup
- Xem lại [Ví dụ](../examples/) cho các mẫu sử dụng
- Khám phá [Hướng dẫn Production](../production.md) cho best practices triển khai
