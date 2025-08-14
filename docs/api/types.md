# Types & Interfaces Reference

This document provides comprehensive information about all types, interfaces, and data structures used in BunnyHop.

## Core Types

### LoadBalanceStrategy

Defines the load balancing strategy for connection pools.

```go
type LoadBalanceStrategy int

const (
    RoundRobin         LoadBalanceStrategy = iota // Distribute requests evenly
    Random                                       // Randomly select nodes
    LeastUsed                                   // Select least used node
    WeightedRoundRobin                          // Distribute based on weights
)
```

**Usage:**
```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

### Logger

Interface for custom logging implementations.

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

**Default Implementation:**
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

**Custom Implementation:**
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

Client configuration structure.

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

**Field Details:**
- `URLs`: List of RabbitMQ connection URLs (default: `["amqp://localhost:5672"]`)
- `ReconnectInterval`: Time between reconnection attempts (default: `5s`)
- `MaxReconnectAttempt`: Maximum number of reconnection attempts (default: `10`)
- `DebugLog`: Enable/disable debug logging (default: `false`)
- `Logger`: Custom logger implementation (default: `DefaultLogger`)
- `TLSConfig`: TLS/SSL configuration (optional)

**Example:**
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

Pool configuration structure.

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

**Field Details:**
- `URLs`: List of RabbitMQ node URLs (required)
- `ReconnectInterval`: Time between reconnection attempts (default: `5s`)
- `MaxReconnectAttempt`: Maximum number of reconnection attempts (default: `10`)
- `HealthCheckInterval`: Interval between health checks (default: `30s`)
- `LoadBalanceStrategy`: Load balancing strategy (default: `RoundRobin`)
- `DebugLog`: Enable/disable debug logging (default: `false`)
- `Logger`: Custom logger implementation (default: `DefaultLogger`)
- `TLSConfig`: TLS/SSL configuration (optional)

**Example:**
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

Pool statistics and health information.

```go
type PoolStats struct {
    TotalNodes          int                // Total number of nodes
    HealthyNodes        int                // Number of healthy nodes
    UnhealthyNodes      int                // Number of unhealthy nodes
    LastHealthCheck     time.Time          // Last health check timestamp
    LoadBalanceStrategy LoadBalanceStrategy // Current load balancing strategy
}
```

**Usage:**
```go
stats := pool.GetStats()
log.Printf("Total nodes: %d", stats.TotalNodes)
log.Printf("Healthy nodes: %d", stats.HealthyNodes)
log.Printf("Unhealthy nodes: %d", stats.UnhealthyNodes)
log.Printf("Last health check: %v", stats.LastHealthCheck)
log.Printf("Strategy: %v", stats.LoadBalanceStrategy)
```

### ConnectionStats

Individual connection statistics.

```go
type ConnectionStats struct {
    URL           string    // Connection URL
    IsConnected   bool      // Connection status
    LastUsed      time.Time // Last usage timestamp
    UsageCount    int64     // Total usage count
    LastError     error     // Last error encountered
    ReconnectAttempts int   // Number of reconnection attempts
}
```

## Internal Types

### NodeConnection

Internal node connection management.

```go
type NodeConnection struct {
    URL           string         // Node URL
    Client        *Client        // Client instance
    IsHealthy     bool           // Health status
    LastUsed      time.Time      // Last usage timestamp
    UsageCount    int64          // Usage counter
    Weight        int            // Node weight (for weighted strategies)
    mu            sync.RWMutex   // Mutex for thread safety
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

Connection state enumeration.

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

BunnyHop defines several custom error types for better error handling.

```go
var (
    ErrNoHealthyNodes    = errors.New("no healthy nodes available")
    ErrPoolNotStarted    = errors.New("pool not started")
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrConnectionFailed  = errors.New("connection failed")
    ErrReconnectFailed   = errors.New("reconnection failed")
)
```

**Error Handling:**
```go
client, err := pool.GetClient()
if err != nil {
    switch {
    case errors.Is(err, bunnyhop.ErrNoHealthyNodes):
        log.Fatal("No healthy nodes available")
    case errors.Is(err, bunnyhop.ErrPoolNotStarted):
        log.Fatal("Pool not started")
    case errors.Is(err, bunnyhop.ErrInvalidConfig):
        log.Fatal("Invalid configuration")
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Utility Types

### URLParser

Utility for parsing and validating RabbitMQ URLs.

```go
type URLParser struct {
    Scheme   string // amqp or amqps
    Username string // Username
    Password string // Password
    Host     string // Hostname
    Port     int    // Port number
    VHost    string // Virtual host
}
```

**Usage:**
```go
parser := &URLParser{}
if err := parser.Parse("amqp://user:pass@host:5672/vhost"); err != nil {
    log.Printf("Invalid URL: %v", err)
    return
}

log.Printf("Host: %s, Port: %d, VHost: %s", parser.Host, parser.Port, parser.VHost)
```

### HealthChecker

Interface for custom health checking implementations.

```go
type HealthChecker interface {
    CheckHealth(client *Client) bool
    GetHealthMetrics() HealthMetrics
}
```

**Default Implementation:**
```go
type DefaultHealthChecker struct {
    timeout time.Duration
}

func (h *DefaultHealthChecker) CheckHealth(client *Client) bool {
    if !client.IsConnected() {
        return false
    }
    
    // Perform basic health check
    ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
    defer cancel()
    
    // Try to get a channel
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

Always use the defined types instead of raw values:

```go
// Good
config := bunnyhop.Config{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}

// Avoid
config := bunnyhop.Config{
    LoadBalanceStrategy: 0, // Magic number
}
```

### 2. Interface Implementation

Implement interfaces correctly:

```go
type CustomLogger struct {
    logger *zap.Logger
}

// Ensure all methods are implemented
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

Use custom error types for better error handling:

```go
if err := pool.Start(); err != nil {
    if errors.Is(err, bunnyhop.ErrInvalidConfig) {
        log.Fatal("Configuration error, check your settings")
    }
    log.Fatalf("Failed to start pool: %v", err)
}
```

### 4. Configuration Validation

Validate configuration before use:

```go
func validateConfig(config bunnyhop.Config) error {
    if len(config.URLs) == 0 {
        return errors.New("at least one URL is required")
    }
    
    if config.ReconnectInterval <= 0 {
        return errors.New("reconnect interval must be positive")
    }
    
    if config.MaxReconnectAttempt < 0 {
        return errors.New("max reconnect attempts cannot be negative")
    }
    
    return nil
}
```

## Examples

### Custom Logger Implementation

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
    // ... use client
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
        return true // Skip check if too recent
    }
    
    h.lastCheck = now
    
    // Perform custom health check
    if !client.IsConnected() {
        return false
    }
    
    // Additional custom checks
    return true
}

func (h *CustomHealthChecker) GetHealthMetrics() HealthMetrics {
    return HealthMetrics{
        LastCheck: h.lastCheck,
        Status:    "healthy",
    }
}
```

## Next Steps

- Read [Client API Reference](client.md) for client operations
- Check [Pool API Reference](pool.md) for pool management
- Review [Configuration Guide](../configuration.md) for usage examples
- Explore [Examples](../examples/) for complete working examples
