# Hướng dẫn cài đặt

Hướng dẫn này sẽ giúp bạn cài đặt và thiết lập BunnyHop trong dự án Go của mình.

## Yêu cầu hệ thống

Trước khi cài đặt BunnyHop, hãy đảm bảo bạn có:

- **Go 1.24 trở lên** - [Tải Go](https://golang.org/dl/)
- **Git** - Để clone repository
- **Docker & Docker Compose** - Để chạy RabbitMQ (tùy chọn nhưng khuyến nghị)

## Phương pháp cài đặt

### Phương pháp 1: Go Modules (Khuyến nghị)

Thêm BunnyHop vào Go module của bạn:

```bash
go get github.com/vanduc0209/bunnyhop
```

Sau đó import trong code Go:

```go
import "github.com/vanduc0209/bunnyhop"
```

### Phương pháp 2: Clone Repository

Clone repository và sử dụng cục bộ:

```bash
git clone https://github.com/vanduc0209/bunnyhop.git
cd bunnyhop
go mod download
```

## Dependencies

BunnyHop có các dependencies sau:

- **github.com/rabbitmq/amqp091-go** - Triển khai giao thức AMQP 0.9.1
- **github.com/stretchr/testify** - Framework testing (cho development)

Các dependencies này sẽ được tải tự động khi bạn chạy `go mod download`.

## Thiết lập RabbitMQ

### Tùy chọn 1: Docker (Khuyến nghị cho Development)

#### Thiết lập Single Node

```bash
# Khởi động một node RabbitMQ
docker-compose -f docker/docker-compose.single.yml up -d

# Hoặc sử dụng script có sẵn
chmod +x scripts/start-single.sh
./scripts/start-single.sh
```

#### Thiết lập Cluster

```bash
# Khởi động cluster 3 node
docker-compose -f docker/docker-compose.cluster.yml up -d

# Hoặc sử dụng script có sẵn
chmod +x scripts/start-cluster.sh
./scripts/start-cluster.sh
```

### Tùy chọn 2: Cài đặt cục bộ

1. Tải RabbitMQ từ [rabbitmq.com](https://www.rabbitmq.com/download.html)
2. Cài đặt và khởi động service
3. Bật management plugin: `rabbitmq-plugins enable rabbitmq_management`

### Tùy chọn 3: Cloud Services

- **RabbitMQ Cloud** - [cloud.rabbitmq.com](https://www.cloudamqp.com/)
- **AWS MQ** - Dịch vụ RabbitMQ được quản lý
- **Azure Service Bus** - Dịch vụ messaging thay thế

## Xác minh cài đặt

### Kiểm tra cài đặt

Tạo file test đơn giản để xác minh mọi thứ hoạt động:

```go
package main

import (
    "log"
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    config := bunnyhop.Config{
        URLs: []string{"amqp://guest:guest@localhost:5672/"},
        DebugLog: true,
    }
    
    client := bunnyhop.NewClient(config)
    log.Printf("BunnyHop client được tạo thành công: %v", client)
}
```

Chạy test:

```bash
go run main.go
```

### Kiểm tra RabbitMQ

Nếu sử dụng Docker, xác minh RabbitMQ đang chạy:

```bash
# Kiểm tra trạng thái container
docker ps

# Truy cập management UI
# Single node: http://localhost:15672 (guest/guest)
# Cluster: http://localhost:15670 (qua HAProxy)
```

## Cấu hình

### Biến môi trường

Thiết lập các biến môi trường này cho production:

```bash
export RABBITMQ_URL="amqp://user:pass@host:5672/"
export RABBITMQ_DEBUG="true"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
```

### File cấu hình

Tạo file cấu hình cho ứng dụng của bạn:

```yaml
# config.yaml
rabbitmq:
  urls:
    - "amqp://user:pass@node1:5672/"
    - "amqp://user:pass@node2:5672/"
  reconnect_interval: "5s"
  max_reconnect_attempts: 10
  debug_log: false
```

## Xử lý sự cố

### Các vấn đề thường gặp

1. **Connection Refused**
   - Đảm bảo RabbitMQ đang chạy
   - Kiểm tra khả năng truy cập port
   - Xác minh thông tin đăng nhập

2. **Import Errors**
   - Chạy `go mod tidy`
   - Kiểm tra tương thích phiên bản Go
   - Xác minh đường dẫn module

3. **Docker Issues**
   - Đảm bảo Docker đang chạy
   - Kiểm tra xung đột port
   - Xác minh phiên bản Docker Compose

### Tìm kiếm trợ giúp

- Kiểm tra [Hướng dẫn xử lý sự cố](troubleshooting.md)
- Xem lại [Tài liệu RabbitMQ](https://www.rabbitmq.com/documentation.html)
- Mở issue trên [GitHub](https://github.com/vanduc0209/bunnyhop/issues)

## Bước tiếp theo

Bây giờ bạn đã cài đặt BunnyHop:

1. Đọc [Hướng dẫn nhanh](quickstart.md) để bắt đầu sử dụng thư viện
2. Khám phá [Ví dụ cơ bản](examples/basic.md) cho các trường hợp sử dụng phổ biến
3. Xem lại [Tùy chọn cấu hình](configuration.md) cho thiết lập nâng cao
4. Kiểm tra [Triển khai Production](production.md) cho best practices 