# Hướng dẫn nhanh

Bắt đầu sử dụng BunnyHop trong vài phút! Hướng dẫn này sẽ chỉ cho bạn cách tạo kết nối RabbitMQ đầu tiên và bắt đầu gửi tin nhắn.

## Yêu cầu

- Go 1.24+ đã cài đặt
- RabbitMQ đang chạy (xem [Hướng dẫn cài đặt](installation.md))
- Hiểu biết cơ bản về Go

## Kết nối đầu tiên của bạn

### 1. Tạo Client đơn giản

Tạo file Go mới và thêm code này:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    // Tạo cấu hình client
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    // Tạo client mới
    client := bunnyhop.NewClient(config)
    
    // Kết nối đến RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Kết nối thất bại: %v", err)
    }
    
    log.Println("Kết nối thành công đến RabbitMQ!")
    
    // Đừng quên đóng kết nối
    defer client.Close()
}
```

### 2. Chạy code của bạn

```bash
go run main.go
```

Bạn sẽ thấy: `Kết nối thành công đến RabbitMQ!`

## Các thao tác cơ bản

### Khai báo Queue

```go
// Khai báo queue bền vững
queue, err := client.DeclareQueue("my_queue", true, false, false, nil)
if err != nil {
    log.Printf("Khai báo queue thất bại: %v", err)
} else {
    log.Printf("Queue đã được khai báo: %s", queue.Name)
}
```

### Khai báo Exchange

```go
// Khai báo exchange trực tiếp
err = client.DeclareExchange("my_exchange", "direct", true, false, false, nil)
if err != nil {
    log.Printf("Khai báo exchange thất bại: %v", err)
} else {
    log.Println("Exchange đã được khai báo thành công")
}
```

### Liên kết Queue với Exchange

```go
// Liên kết queue với exchange bằng routing key
err = client.QueueBind("my_queue", "my_key", "my_exchange", false, nil)
if err != nil {
    log.Printf("Liên kết queue thất bại: %v", err)
} else {
    log.Println("Queue đã được liên kết thành công")
}
```

### Gửi tin nhắn

```go
// Tạo tin nhắn
message := amqp.Publishing{
    ContentType: "text/plain",
    Body:        []byte("Xin chào, RabbitMQ!"),
}

// Gửi đến exchange
err = client.PublishMessage("my_exchange", "my_key", false, false, message)
if err != nil {
    log.Printf("Gửi tin nhắn thất bại: %v", err)
} else {
    log.Println("Tin nhắn đã được gửi thành công")
}
```

## Sử dụng Connection Pool

Đối với ứng dụng production, sử dụng connection pool để có độ tin cậy tốt hơn:

### 1. Tạo cấu hình Pool

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
```

### 2. Tạo và khởi động Pool

```go
// Tạo pool mới
pool := bunnyhop.NewPool(config)

// Khởi động pool
if err := pool.Start(); err != nil {
    log.Fatalf("Khởi động pool thất bại: %v", err)
}

// Đừng quên đóng pool
defer pool.Close()
```

### 3. Lấy Client từ Pool

```go
// Lấy client sử dụng chiến lược load balancing
client, err := pool.GetClient()
if err != nil {
    log.Printf("Lấy client thất bại: %v", err)
    return
}

// Sử dụng client như trước
queue, err := client.DeclareQueue("pool_queue", true, false, false, nil)
if err != nil {
    log.Printf("Khai báo queue thất bại: %v", err)
} else {
    log.Printf("Queue đã được khai báo: %s", queue.Name)
}
```

## Ví dụ hoàn chỉnh

Đây là ví dụ hoàn chỉnh có thể chạy được:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/rabbitmq/amqp091-go"
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    // Tạo cấu hình pool
    config := bunnyhop.PoolConfig{
        URLs: []string{
            "amqp://guest:guest@localhost:5672/",
        },
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        HealthCheckInterval: 30 * time.Second,
        LoadBalanceStrategy: bunnyhop.RoundRobin,
        DebugLog:            true,
    }
    
    // Tạo và khởi động pool
    pool := bunnyhop.NewPool(config)
    if err := pool.Start(); err != nil {
        log.Fatalf("Khởi động pool thất bại: %v", err)
    }
    defer pool.Close()
    
    // Chờ pool sẵn sàng
    time.Sleep(2 * time.Second)
    
    // Lấy client từ pool
    client, err := pool.GetClient()
    if err != nil {
        log.Printf("Lấy client thất bại: %v", err)
        return
    }
    
    // Khai báo queue
    queue, err := client.DeclareQueue("hello", true, false, false, nil)
    if err != nil {
        log.Printf("Khai báo queue thất bại: %v", err)
        return
    }
    log.Printf("Queue đã được khai báo: %s", queue.Name)
    
    // Khai báo exchange
    err = client.DeclareExchange("hello_exchange", "direct", true, false, false, nil)
    if err != nil {
        log.Printf("Khai báo exchange thất bại: %v", err)
        return
    }
    
    // Liên kết queue với exchange
    err = client.QueueBind("hello", "hello_key", "hello_exchange", false, nil)
    if err != nil {
        log.Printf("Liên kết queue thất bại: %v", err)
        return
    }
    
    // Gửi tin nhắn
    message := amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte("Xin chào, BunnyHop!"),
    }
    
    err = client.PublishMessage("hello_exchange", "hello_key", false, false, message)
    if err != nil {
        log.Printf("Gửi tin nhắn thất bại: %v", err)
        return
    }
    
    log.Println("Tin nhắn đã được gửi thành công!")
    
    // Lấy thống kê pool
    stats := pool.GetStats()
    log.Printf("Thống kê pool: %+v", stats)
}
```

## Bước tiếp theo

Bây giờ bạn đã có kiến thức cơ bản:

1. **Khám phá Ví dụ**: Xem [Ví dụ cơ bản](examples/basic.md) để biết thêm các mẫu
2. **Tìm hiểu Cấu hình**: Đọc [Hướng dẫn cấu hình](configuration.md) cho các tùy chọn nâng cao
3. **Hiểu Kiến trúc**: Xem [Tổng quan kiến trúc](architecture.md) để biết cách hoạt động
4. **Sẵn sàng Production**: Xem lại [Triển khai Production](production.md) cho best practices

## Cần giúp đỡ?

- Kiểm tra [Hướng dẫn xử lý sự cố](troubleshooting.md)
- Xem lại [Tham chiếu API](api/) để biết tài liệu chi tiết về các method
- Mở issue trên [GitHub](https://github.com/vanduc0209/bunnyhop/issues) 