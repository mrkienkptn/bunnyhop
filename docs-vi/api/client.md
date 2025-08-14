# Tham chiếu API Client

Type `Client` cung cấp một kết nối đơn lẻ đến RabbitMQ với khả năng tự động kết nối lại.

## Tổng quan

```go
type Client struct {
    // Các trường private để sử dụng nội bộ
}
```

## Constructor

### NewClient

Tạo một RabbitMQ client mới với cấu hình được chỉ định.

```go
func NewClient(config Config) *Client
```

**Tham số:**
- `config` - Cấu hình client

**Trả về:**
- `*Client` - Instance client mới

**Ví dụ:**
```go
config := bunnyhop.Config{
    URLs:                []string{"amqp://guest:guest@localhost:5672/"},
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    DebugLog:            true,
}

client := bunnyhop.NewClient(config)
```

## Methods

### Connect

Thiết lập kết nối đến RabbitMQ.

```go
func (c *Client) Connect(ctx context.Context) error
```

**Tham số:**
- `ctx` - Context cho thao tác kết nối

**Trả về:**
- `error` - Lỗi kết nối nếu có

**Ví dụ:**
```go
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatalf("Kết nối thất bại: %v", err)
}
```

### IsConnected

Kiểm tra xem client có đang kết nối đến RabbitMQ hay không.

```go
func (c *Client) IsConnected() bool
```

**Trả về:**
- `bool` - True nếu đã kết nối, false nếu không

**Ví dụ:**
```go
if client.IsConnected() {
    log.Println("Client đã kết nối")
} else {
    log.Println("Client chưa kết nối")
}
```

### GetChannel

Lấy AMQP channel hiện tại.

```go
func (c *Client) GetChannel() (*amqp.Channel, error)
```

**Trả về:**
- `*amqp.Channel` - AMQP channel hiện tại
- `error` - Lỗi nếu client chưa kết nối

**Ví dụ:**
```go
channel, err := client.GetChannel()
if err != nil {
    log.Printf("Lấy channel thất bại: %v", err)
    return
}

// Sử dụng channel cho các thao tác AMQP
```

### GetConnection

Lấy AMQP connection hiện tại.

```go
func (c *Client) GetConnection() (*amqp.Connection, error)
```

**Trả về:**
- `*amqp.Connection` - AMQP connection hiện tại
- `error` - Lỗi nếu client chưa kết nối

**Ví dụ:**
```go
connection, err := client.GetConnection()
if err != nil {
    log.Printf("Lấy connection thất bại: %v", err)
    return
}

// Sử dụng connection cho các thao tác nâng cao
```

### Close

Đóng kết nối client và dọn dẹp tài nguyên.

```go
func (c *Client) Close() error
```

**Trả về:**
- `error` - Bất kỳ lỗi nào gặp phải trong quá trình dọn dẹp

**Ví dụ:**
```go
defer func() {
    if err := client.Close(); err != nil {
        log.Printf("Lỗi đóng client: %v", err)
    }
}()
```

## Các thao tác RabbitMQ

### PublishMessage

Gửi tin nhắn đến exchange.

```go
func (c *Client) PublishMessage(
    exchange, routingKey string,
    mandatory, immediate bool,
    msg amqp.Publishing,
) error
```

**Tham số:**
- `exchange` - Tên exchange
- `routingKey` - Routing key cho tin nhắn
- `mandatory` - Tin nhắn có bắt buộc hay không
- `immediate` - Tin nhắn có ngay lập tức hay không
- `msg` - Tin nhắn để gửi

**Trả về:**
- `error` - Lỗi gửi tin nhắn nếu có

**Ví dụ:**
```go
message := amqp.Publishing{
    ContentType: "text/plain",
    Body:        []byte("Xin chào, RabbitMQ!"),
}

err := client.PublishMessage("my_exchange", "my_key", false, false, message)
if err != nil {
    log.Printf("Gửi tin nhắn thất bại: %v", err)
}
```

### DeclareQueue

Khai báo queue trên RabbitMQ.

```go
func (c *Client) DeclareQueue(
    name string,
    durable, autoDelete, exclusive bool,
    args amqp.Table,
) (amqp.Queue, error)
```

**Tham số:**
- `name` - Tên queue
- `durable` - Queue có tồn tại sau khi restart broker hay không
- `autoDelete` - Queue có bị xóa khi consumer cuối cùng hủy đăng ký hay không
- `exclusive` - Queue có độc quyền cho connection này hay không
- `args` - Các tham số queue bổ sung

**Trả về:**
- `amqp.Queue` - Thông tin queue đã khai báo
- `error` - Lỗi khai báo nếu có

**Ví dụ:**
```go
queue, err := client.DeclareQueue("my_queue", true, false, false, nil)
if err != nil {
    log.Printf("Khai báo queue thất bại: %v", err)
    return
}

log.Printf("Queue đã được khai báo: %s", queue.Name)
```

### DeclareExchange

Khai báo exchange trên RabbitMQ.

```go
func (c *Client) DeclareExchange(
    name, kind string,
    durable, autoDelete, internal bool,
    args amqp.Table,
) error
```

**Tham số:**
- `name` - Tên exchange
- `kind` - Loại exchange (direct, fanout, topic, headers)
- `durable` - Exchange có tồn tại sau khi restart broker hay không
- `autoDelete` - Exchange có bị xóa khi queue cuối cùng hủy đăng ký hay không
- `internal` - Exchange có phải là internal hay không
- `args` - Các tham số exchange bổ sung

**Trả về:**
- `error` - Lỗi khai báo nếu có

**Ví dụ:**
```go
err := client.DeclareExchange("my_exchange", "direct", true, false, false, nil)
if err != nil {
    log.Printf("Khai báo exchange thất bại: %v", err)
    return
}

log.Println("Exchange đã được khai báo thành công")
```

### QueueBind

Liên kết queue với exchange bằng routing key.

```go
func (c *Client) QueueBind(
    name, key, exchange string,
    noWait bool,
    args amqp.Table,
) error
```

**Tham số:**
- `name` - Tên queue
- `key` - Routing key
- `exchange` - Tên exchange
- `noWait` - Có chờ xác nhận từ server hay không
- `args` - Các tham số binding bổ sung

**Trả về:**
- `error` - Lỗi binding nếu có

**Ví dụ:**
```go
err := client.QueueBind("my_queue", "my_key", "my_exchange", false, nil)
if err != nil {
    log.Printf("Liên kết queue thất bại: %v", err)
    return
}

log.Println("Queue đã được liên kết thành công")
```

## Cấu hình

### Config

Cấu trúc cấu hình client.

```go
type Config struct {
    URLs                []string      // Danh sách URL kết nối RabbitMQ
    ReconnectInterval   time.Duration // Thời gian giữa các lần thử kết nối lại
    MaxReconnectAttempt int           // Số lần thử kết nối lại tối đa
    DebugLog            bool          // Bật/tắt debug logging
    Logger              Logger        // Triển khai logger tùy chỉnh
}
```

**Các trường:**
- `URLs` - Danh sách URL kết nối RabbitMQ (mặc định: `["amqp://localhost:5672"]`)
- `ReconnectInterval` - Thời gian giữa các lần thử kết nối lại (mặc định: `5s`)
- `MaxReconnectAttempt` - Số lần thử kết nối lại tối đa (mặc định: `10`)
- `DebugLog` - Bật/tắt debug logging (mặc định: `false`)
- `Logger` - Triển khai logger tùy chỉnh (mặc định: `DefaultLogger`)

## Xử lý lỗi

Client tự động xử lý các lỗi kết nối và thử kết nối lại dựa trên cấu hình. Các tình huống lỗi thường gặp:

- **Connection refused** - Server RabbitMQ không chạy
- **Authentication failed** - Thông tin đăng nhập không hợp lệ
- **Channel error** - Thao tác AMQP channel thất bại
- **Network timeout** - Kết nối bị timeout

## Best Practices

1. **Luôn đóng client** sử dụng `defer client.Close()`
2. **Kiểm tra trạng thái kết nối** trước khi thực hiện các thao tác sử dụng `IsConnected()`
3. **Xử lý lỗi một cách nhẹ nhàng** và triển khai logic retry cho các thao tác quan trọng
4. **Sử dụng context** cho các thao tác kết nối để hỗ trợ hủy bỏ
5. **Monitor sức khỏe kết nối** trong môi trường production

## Ví dụ

Xem [Ví dụ cơ bản](../examples/basic.md) và [Ví dụ nâng cao](../examples/advanced.md) để biết các ví dụ hoàn chỉnh có thể chạy được. 