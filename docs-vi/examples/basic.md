# Ví dụ cơ bản

Phần này cung cấp các ví dụ cơ bản về các thao tác RabbitMQ phổ biến sử dụng BunnyHop.

## Gửi tin nhắn đơn giản

### Publisher cơ bản

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
    // Tạo cấu hình client
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    // Tạo client mới
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Kết nối đến RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Kết nối thất bại: %v", err)
    }
    
    // Khai báo exchange
    err := client.DeclareExchange("test_exchange", "direct", true, false, false, nil)
    if err != nil {
        log.Fatalf("Khai báo exchange thất bại: %v", err)
    }
    
    // Khai báo queue
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Khai báo queue thất bại: %v", err)
    }
    
    // Liên kết queue với exchange
    err = client.QueueBind("test_queue", "test_key", "test_exchange", false, nil)
    if err != nil {
        log.Fatalf("Liên kết queue thất bại: %v", err)
    }
    
    // Gửi tin nhắn
    message := amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte("Xin chào, RabbitMQ!"),
        Timestamp:   time.Now(),
    }
    
    err = client.PublishMessage("test_exchange", "test_key", false, false, message)
    if err != nil {
        log.Fatalf("Gửi tin nhắn thất bại: %v", err)
    }
    
    log.Printf("Tin nhắn đã được gửi đến queue: %s", queue.Name)
}
```

## Consumer tin nhắn

### Consumer cơ bản

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
    // Tạo cấu hình client
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    // Tạo client mới
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Kết nối đến RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Kết nối thất bại: %v", err)
    }
    
    // Lấy channel cho các thao tác consumer
    channel, err := client.GetChannel()
    if err != nil {
        log.Fatalf("Lấy channel thất bại: %v", err)
    }
    
    // Khai báo queue (giống như publisher)
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Khai báo queue thất bại: %v", err)
    }
    
    // Thiết lập QoS để phân phối công bằng
    err = channel.Qos(1, 0, false)
    if err != nil {
        log.Fatalf("Thiết lập QoS thất bại: %v", err)
    }
    
    // Bắt đầu tiêu thụ tin nhắn
    msgs, err := channel.Consume(
        queue.Name, // queue
        "",         // consumer
        false,      // auto-ack
        false,      // exclusive
        false,      // no-local
        false,      // no-wait
        nil,        // args
    )
    if err != nil {
        log.Fatalf("Bắt đầu tiêu thụ thất bại: %v", err)
    }
    
    log.Printf("Bắt đầu tiêu thụ từ queue: %s", queue.Name)
    
    // Xử lý tin nhắn
    for msg := range msgs {
        log.Printf("Nhận tin nhắn: %s", string(msg.Body))
        
        // Xử lý tin nhắn ở đây
        time.Sleep(100 * time.Millisecond) // Mô phỏng xử lý
        
        // Xác nhận tin nhắn
        msg.Ack(false)
    }
}
```

## Sử dụng Connection Pool

### Publisher dựa trên Pool

```go
package main

import (
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
            "amqp://guest:guest@localhost:5673/",
        },
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        HealthCheckInterval: 30 * time.Second,
        LoadBalanceStrategy: bunnyhop.RoundRobin,
        DebugLog:            true,
    }
    
    // Tạo pool mới
    pool := bunnyhop.NewPool(config)
    defer pool.Close()
    
    // Khởi động pool
    if err := pool.Start(); err != nil {
        log.Fatalf("Khởi động pool thất bại: %v", err)
    }
    
    // Chờ pool sẵn sàng
    time.Sleep(2 * time.Second)
    
    // Lấy client từ pool
    client, err := pool.GetClient()
    if err != nil {
        log.Fatalf("Lấy client thất bại: %v", err)
    }
    
    // Sử dụng client cho các thao tác
    queue, err := client.DeclareQueue("pool_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Khai báo queue thất bại: %v", err)
    }
    
    // Gửi nhiều tin nhắn
    for i := 0; i < 10; i++ {
        message := amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(fmt.Sprintf("Tin nhắn %d từ pool", i)),
        }
        
        err = client.PublishMessage("", queue.Name, false, false, message)
        if err != nil {
            log.Printf("Gửi tin nhắn %d thất bại: %v", i, err)
        } else {
            log.Printf("Tin nhắn %d đã được gửi thành công", i)
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Lấy thống kê pool
    stats := pool.GetStats()
    log.Printf("Thống kê pool: %+v", stats)
}
```

## Xử lý lỗi

### Xử lý lỗi mạnh mẽ

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Thử kết nối lại với exponential backoff
    maxRetries := 5
    for attempt := 0; attempt < maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        
        err := client.Connect(ctx)
        cancel()
        
        if err == nil {
            log.Println("Kết nối thành công đến RabbitMQ")
            break
        }
        
        log.Printf("Lần thử kết nối %d thất bại: %v", attempt+1, err)
        
        if attempt < maxRetries-1 {
            backoff := time.Duration(attempt+1) * time.Second
            log.Printf("Thử lại sau %v...", backoff)
            time.Sleep(backoff)
        } else {
            log.Fatalf("Kết nối thất bại sau %d lần thử", maxRetries)
        }
    }
    
    // Kiểm tra sức khỏe kết nối
    if !client.IsConnected() {
        log.Fatal("Client không được kết nối")
    }
    
    log.Println("Sẵn sàng thực hiện các thao tác")
}
```

## Ví dụ cấu hình

### Cấu hình dựa trên môi trường

```go
package main

import (
    "log"
    "os"
    "strconv"
    "time"
    
    "github.com/vanduc0209/bunnyhop"
)

func getConfig() bunnyhop.Config {
    // Lấy cấu hình từ biến môi trường
    urls := []string{"amqp://localhost:5672"} // Mặc định
    if envURLs := os.Getenv("RABBITMQ_URLS"); envURLs != "" {
        urls = strings.Split(envURLs, ",")
    }
    
    reconnectInterval := 5 * time.Second // Mặc định
    if envInterval := os.Getenv("RABBITMQ_RECONNECT_INTERVAL"); envInterval != "" {
        if interval, err := time.ParseDuration(envInterval); err == nil {
            reconnectInterval = interval
        }
    }
    
    maxAttempts := 10 // Mặc định
    if envAttempts := os.Getenv("RABBITMQ_MAX_RECONNECT_ATTEMPTS"); envAttempts != "" {
        if attempts, err := strconv.Atoi(envAttempts); err == nil {
            maxAttempts = attempts
        }
    }
    
    debugLog := false // Mặc định
    if envDebug := os.Getenv("RABBITMQ_DEBUG"); envDebug != "" {
        debugLog = envDebug == "true"
    }
    
    return bunnyhop.Config{
        URLs:                urls,
        ReconnectInterval:   reconnectInterval,
        MaxReconnectAttempt: maxAttempts,
        DebugLog:            debugLog,
    }
}

func main() {
    config := getConfig()
    log.Printf("Sử dụng cấu hình: %+v", config)
    
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Sử dụng client...
}
```

## Chạy các ví dụ

1. **Khởi động RabbitMQ** (sử dụng Docker):
   ```bash
   make docker-single
   ```

2. **Chạy publisher**:
   ```bash
   go run publisher.go
   ```

3. **Chạy consumer** (trong terminal khác):
   ```bash
   go run consumer.go
   ```

## Bước tiếp theo

- Khám phá [Ví dụ nâng cao](advanced.md) cho các tình huống phức tạp
- Kiểm tra [Ví dụ tích hợp](integration.md) cho các trường hợp sử dụng thực tế
- Xem lại [Tham chiếu API](../api/) để biết tài liệu hoàn chỉnh về các method
- Đọc [Triển khai Production](../production.md) cho best practices 