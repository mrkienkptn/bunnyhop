# Configuration Guide

This guide covers all available configuration options for BunnyHop and how to use them effectively.

## Configuration Overview

BunnyHop provides two main configuration types:

1. **Client Config** - For single connections
2. **Pool Config** - For connection pools with multiple nodes

## Client Configuration

### Config Structure

```go
type Config struct {
    URLs                []string      // List of RabbitMQ connection URLs
    ReconnectInterval   time.Duration // Time between reconnection attempts
    MaxReconnectAttempt int           // Maximum number of reconnection attempts
    DebugLog            bool          // Enable/disable debug logging
    Logger              Logger        // Custom logger implementation
    TLSConfig           *tls.Config   // TLS/SSL configuration
}
```

### Basic Options

```go
config := bunnyhop.Config{
    URLs:                []string{"amqp://guest:guest@localhost:5672/"},
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    DebugLog:            true,
}
```

### Advanced Configuration

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

### PoolConfig Structure

```go
type PoolConfig struct {
    URLs                []string           // List of RabbitMQ node URLs
    ReconnectInterval   time.Duration      // Time between reconnection attempts
    MaxReconnectAttempt int                // Maximum number of reconnection attempts
    HealthCheckInterval time.Duration      // Health check interval
    LoadBalanceStrategy LoadBalanceStrategy // Load balancing strategy
    DebugLog            bool               // Enable/disable debug logging
    Logger              Logger             // Custom logger implementation
    TLSConfig           *tls.Config        // TLS/SSL configuration
}
```

### Basic Pool Configuration

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

### 1. Round Robin (Default)

Distributes requests evenly across all healthy nodes in sequence.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

**Advantages:**
- Even distribution
- Simple and easy to understand
- Suitable for most cases

**Disadvantages:**
- Doesn't consider individual node capacity
- May overload weaker nodes

### 2. Random

Randomly selects a healthy node for each request.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.Random,
}
```

**Advantages:**
- Random distribution
- Good for high-throughput scenarios
- Minimizes predictable patterns

**Disadvantages:**
- Doesn't guarantee even distribution
- May cause local overload

### 3. Least Used

Selects the node with the lowest usage count.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.LeastUsed,
}
```

**Advantages:**
- Optimizes resource usage
- Suitable for resource-intensive operations
- Automatic load balancing

**Disadvantages:**
- Requires tracking usage state
- May cause fluctuations

### 4. Weighted Round Robin

Distributes requests based on node weights.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.WeightedRoundRobin,
}

// Set weights for different node capacities
pool.SetNodeWeight("amqp://node1:5672/", 3)  // High capacity
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Medium capacity
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Lower capacity
```

**Advantages:**
- Flexible and customizable
- Suitable for nodes with different capacities
- Good control over distribution

**Disadvantages:**
- Requires manual configuration
- More complex than other strategies

## TLS/SSL Configuration

### Enable TLS

```go
// Create TLS configuration
tlsConfig := &tls.Config{
    RootCAs:            rootCAs,
    Certificates:       []tls.Certificate{cert},
    InsecureSkipVerify: false,
}

// Use in configuration
config := bunnyhop.Config{
    URLs:      []string{"amqps://user:pass@node1:5671/"},
    TLSConfig: tlsConfig,
}
```

### Using Certificate Files

```go
// Read CA certificate
caCert, err := ioutil.ReadFile("/path/to/ca-cert.pem")
if err != nil {
    log.Fatal(err)
}

caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

// Read client certificate
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

## Custom Logger Configuration

### Logger Interface

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

### Custom Logger Implementation

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

// Use in configuration
customLogger := &CustomLogger{logger: setupZapLogger()}
config := bunnyhop.Config{
    URLs:   []string{"amqp://localhost:5672/"},
    Logger: customLogger,
}
```

## Environment-based Configuration

### Basic Environment Variables

```bash
export RABBITMQ_URLS="amqp://user:pass@node1:5672/,amqp://user:pass@node2:5672/"
export RABBITMQ_DEBUG="true"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
```

### Pool Environment Variables

```bash
export RABBITMQ_HEALTH_CHECK_INTERVAL="30s"
export RABBITMQ_LOAD_BALANCE_STRATEGY="WeightedRoundRobin"
```

### Helper Function to Read Configuration

```go
func getConfigFromEnv() bunnyhop.Config {
    urls := []string{"amqp://localhost:5672"} // Default
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

## File-based Configuration

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

### Reading Configuration from File

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

### 1. Environment-based Configuration

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

### 2. Configuration Validation

```go
func validateConfig(config bunnyhop.Config) error {
    if len(config.URLs) == 0 {
        return errors.New("at least one connection URL is required")
    }
    
    if config.ReconnectInterval <= 0 {
        return errors.New("ReconnectInterval must be greater than 0")
    }
    
    if config.MaxReconnectAttempt < 0 {
        return errors.New("MaxReconnectAttempt cannot be negative")
    }
    
    return nil
}
```

### 3. Dynamic Configuration

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

## Next Steps

- Read [Quick Start Guide](quickstart.md) to get started
- Explore [Basic Examples](examples/basic.md) for usage patterns
- Review [Production Deployment](production.md) for best practices
- Check [API Reference](api/) for detailed method documentation 