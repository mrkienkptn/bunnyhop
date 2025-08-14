# Hướng dẫn cấu hình

Hướng dẫn này bao gồm tất cả các tùy chọn cấu hình có sẵn cho BunnyHop và cách sử dụng chúng một cách hiệu quả.

## Tổng quan cấu hình

BunnyHop cung cấp hai loại cấu hình chính:

1. **Client Config** - Cho kết nối đơn lẻ
2. **Pool Config** - Cho connection pool với nhiều node

## Client Configuration

### Cấu trúc Config

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

### Các tùy chọn cơ bản

```go
config := bunnyhop.Config{
    URLs:                []string{"amqp://guest:guest@localhost:5672/"},
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    DebugLog:            true,
}
```

### Cấu hình nâng cao

```go
config := bunnyhop.Config{
    URLs: []string{
        "amqp://user1:pass1@node1:5672/",
        "amqp://user2:pass2@node2:5672/",
    },
    ReconnectInterval:   3 * time.Second,
    MaxReconnectAttempt: 15,
    DebugLog:            false,
    Logger:              &CustomLogger{},
}
```

## Pool Configuration

### Cấu trúc PoolConfig

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

### Cấu hình pool cơ bản

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

## Load Balancing Strategies

### 1. Round Robin (Mặc định)

Phân phối yêu cầu đều đặn qua tất cả các node khỏe mạnh theo thứ tự.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

**Ưu điểm:**
- Phân phối đều đặn
- Đơn giản và dễ hiểu
- Phù hợp cho hầu hết các trường hợp

**Nhược điểm:**
- Không tính đến khả năng của từng node
- Có thể gây quá tải cho node yếu

### 2. Random

Chọn ngẫu nhiên một node khỏe mạnh cho mỗi yêu cầu.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.Random,
}
```

**Ưu điểm:**
- Phân phối ngẫu nhiên
- Tốt cho các tình huống throughput cao
- Giảm thiểu pattern có thể dự đoán

**Nhược điểm:**
- Không đảm bảo phân phối đều đặn
- Có thể gây quá tải cục bộ

### 3. Least Used

Chọn node có số lần sử dụng thấp nhất.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.LeastUsed,
}
```

**Ưu điểm:**
- Tối ưu hóa việc sử dụng tài nguyên
- Phù hợp cho các thao tác tiêu tốn tài nguyên
- Cân bằng tải tự động

**Nhược điểm:**
- Cần theo dõi trạng thái sử dụng
- Có thể gây dao động

### 4. Weighted Round Robin

Phân phối dựa trên trọng số của từng node.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.WeightedRoundRobin,
}

// Thiết lập trọng số cho từng node
pool.SetNodeWeight("amqp://node1:5672/", 3)  // Khả năng cao
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Khả năng trung bình
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Khả năng thấp
```

**Ưu điểm:**
- Linh hoạt và có thể tùy chỉnh
- Phù hợp cho các node có khả năng khác nhau
- Kiểm soát tốt việc phân phối

**Nhược điểm:**
- Cần cấu hình thủ công
- Phức tạp hơn các chiến lược khác

## Cấu hình TLS/SSL

### Bật TLS

```go
// Tạo cấu hình TLS
tlsConfig := &tls.Config{
    RootCAs:            rootCAs,
    Certificates:       []tls.Certificate{cert},
    InsecureSkipVerify: false,
}

// Sử dụng trong cấu hình
config := bunnyhop.Config{
    URLs:      []string{"amqps://user:pass@node1:5671/"},
    TLSConfig: tlsConfig,
}
```

### Sử dụng certificate files

```go
// Đọc CA certificate
caCert, err := ioutil.ReadFile("/path/to/ca-cert.pem")
if err != nil {
    log.Fatal(err)
}

caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// Đọc client certificate
cert, err := tls.LoadX509KeyPair("/path/to/client-cert.pem", "/path/to/client-key.pem")
if err != nil {
    log.Fatal(err)
}

tlsConfig := &tls.Config{
    RootCAs:      caCertPool,
    Certificates: []tls.Certificate{cert},
}

config := bunnyhop.Config{
    URLs:      []string{"amqps://user:pass@node1:5671/"},
    TLSConfig: tlsConfig,
}
```

## Cấu hình Logger tùy chỉnh

### Interface Logger

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

### Triển khai logger tùy chỉnh

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

// Sử dụng trong cấu hình
customLogger := &CustomLogger{logger: setupZapLogger()}
config := bunnyhop.Config{
    URLs:   []string{"amqp://localhost:5672/"},
    Logger: customLogger,
}
```

## Cấu hình từ biến môi trường

### Biến môi trường cơ bản

```bash
export RABBITMQ_URLS="amqp://user:pass@node1:5672/,amqp://user:pass@node2:5672/"
export RABBITMQ_DEBUG="true"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
```

### Biến môi trường cho Pool

```bash
export RABBITMQ_HEALTH_CHECK_INTERVAL="30s"
export RABBITMQ_LOAD_BALANCE_STRATEGY="WeightedRoundRobin"
```

### Hàm helper để đọc cấu hình

```go
func getConfigFromEnv() bunnyhop.Config {
    urls := []string{"amqp://localhost:5672"} // Mặc định
    if envURLs := os.Getenv("RABBITMQ_URLS"); envURLs != "" {
        urls = strings.Split(envURLs, ",")
    }
    
    reconnectInterval := 5 * time.Second
    if envInterval := os.Getenv("RABBITMQ_RECONNECT_INTERVAL"); envInterval != "" {
        if interval, err := time.ParseDuration(envInterval); err == nil {
            reconnectInterval = interval
        }
    }
    
    maxAttempts := 10
    if envAttempts := os.Getenv("RABBITMQ_MAX_RECONNECT_ATTEMPTS"); envAttempts != "" {
        if attempts, err := strconv.Atoi(envAttempts); err == nil {
            maxAttempts = attempts
        }
    }
    
    debugLog := false
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
```

## Cấu hình từ file

### YAML Configuration

```yaml
# config/rabbitmq.yaml
rabbitmq:
  urls:
    - "amqp://user:pass@node1:5672/"
    - "amqp://user:pass@node2:5672/"
  
  client:
    reconnect_interval: "5s"
    max_reconnect_attempts: 10
    debug_log: false
  
  pool:
    health_check_interval: "30s"
    load_balance_strategy: "WeightedRoundRobin"
  
  security:
    tls_enabled: true
    ca_cert_path: "/etc/ssl/certs/ca-certificates.crt"
    client_cert_path: "/etc/ssl/certs/client.crt"
    client_key_path: "/etc/ssl/certs/client.key"
```

### JSON Configuration

```json
{
  "rabbitmq": {
    "urls": [
      "amqp://user:pass@node1:5672/",
      "amqp://user:pass@node2:5672/"
    ],
    "client": {
      "reconnect_interval": "5s",
      "max_reconnect_attempts": 10,
      "debug_log": false
    },
    "pool": {
      "health_check_interval": "30s",
      "load_balance_strategy": "WeightedRoundRobin"
    }
  }
}
```

### Đọc cấu hình từ file

```go
type RabbitMQConfig struct {
    RabbitMQ struct {
        URLs []string `yaml:"urls"`
        Client struct {
            ReconnectInterval   string `yaml:"reconnect_interval"`
            MaxReconnectAttempt int    `yaml:"max_reconnect_attempts"`
            DebugLog            bool   `yaml:"debug_log"`
        } `yaml:"client"`
        Pool struct {
            HealthCheckInterval string `yaml:"health_check_interval"`
            LoadBalanceStrategy string `yaml:"load_balance_strategy"`
        } `yaml:"pool"`
    } `yaml:"rabbitmq"`
}

func loadConfigFromFile(filename string) (*RabbitMQConfig, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config RabbitMQConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## Best Practices

### 1. Cấu hình theo môi trường

```go
func getConfig(environment string) bunnyhop.Config {
    switch environment {
    case "development":
        return bunnyhop.Config{
            URLs:                []string{"amqp://localhost:5672/"},
            ReconnectInterval:   1 * time.Second,
            MaxReconnectAttempt: 5,
            DebugLog:            true,
        }
    case "staging":
        return bunnyhop.Config{
            URLs:                []string{"amqp://staging:5672/"},
            ReconnectInterval:   3 * time.Second,
            MaxReconnectAttempt: 8,
            DebugLog:            false,
        }
    case "production":
        return bunnyhop.Config{
            URLs:                []string{"amqp://prod1:5672/", "amqp://prod2:5672/"},
            ReconnectInterval:   5 * time.Second,
            MaxReconnectAttempt: 10,
            DebugLog:            false,
        }
    default:
        return bunnyhop.Config{
            URLs: []string{"amqp://localhost:5672/"},
        }
    }
}
```

### 2. Validation cấu hình

```go
func validateConfig(config bunnyhop.Config) error {
    if len(config.URLs) == 0 {
        return errors.New("ít nhất phải có một URL kết nối")
    }
    
    if config.ReconnectInterval <= 0 {
        return errors.New("ReconnectInterval phải lớn hơn 0")
    }
    
    if config.MaxReconnectAttempt < 0 {
        return errors.New("MaxReconnectAttempt không được âm")
    }
    
    return nil
}
```

### 3. Cấu hình động

```go
type DynamicConfig struct {
    config bunnyhop.Config
    mu     sync.RWMutex
}

func (dc *DynamicConfig) UpdateConfig(newConfig bunnyhop.Config) {
    dc.mu.Lock()
    defer dc.mu.Unlock()
    dc.config = newConfig
}

func (dc *DynamicConfig) GetConfig() bunnyhop.Config {
    dc.mu.RLock()
    defer dc.mu.RUnlock()
    return dc.config
}
```

## Bước tiếp theo

- Xem [Hướng dẫn nhanh](quickstart.md) để bắt đầu sử dụng
- Khám phá [Ví dụ cơ bản](examples/basic.md) cho các trường hợp sử dụng
- Đọc [Triển khai Production](production.md) cho best practices
- Xem [Tham chiếu API](api/) để biết chi tiết về các method 