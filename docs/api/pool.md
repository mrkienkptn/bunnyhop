# Pool API Reference

The `Pool` type provides a connection pool with multiple RabbitMQ nodes, automatic failover, and load balancing capabilities.

## Overview

```go
type Pool struct {
    // Private fields for internal use
}
```

## Constructor

### NewPool

Creates a new connection pool with the specified configuration.

```go
func NewPool(config PoolConfig) *Pool
```

**Parameters:**
- `config` - Pool configuration

**Returns:**
- `*Pool` - New pool instance

**Example:**
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

Starts the pool and establishes connections to all nodes.

```go
func (p *Pool) Start() error
```

**Returns:**
- `error` - Error if starting the pool fails

**Example:**
```go
if err := pool.Start(); err != nil {
    log.Fatalf("Failed to start pool: %v", err)
}
```

### GetClient

Retrieves a client using the configured load balancing strategy.

```go
func (p *Pool) GetClient() (*Client, error)
```

**Returns:**
- `*Client` - RabbitMQ client instance
- `error` - Error if no healthy nodes are available

**Example:**
```go
client, err := pool.GetClient()
if err != nil {
    log.Printf("Failed to get client: %v", err)
    return
}

// Use the client for operations
queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
```

### GetStats

Retrieves pool statistics and health information.

```go
func (p *Pool) GetStats() PoolStats
```

**Returns:**
- `PoolStats` - Pool statistics structure

**Example:**
```go
stats := pool.GetStats()
log.Printf("Pool stats: %+v", stats)
log.Printf("Healthy nodes: %d", stats.HealthyNodes)
log.Printf("Total nodes: %d", stats.TotalNodes)
```

### Close

Closes the pool and all connections.

```go
func (p *Pool) Close() error
```

**Returns:**
- `error` - Any errors encountered during cleanup

**Example:**
```go
defer func() {
    if err := pool.Close(); err != nil {
        log.Printf("Error closing pool: %v", err)
    }
}()
```

### SetNodeWeight

Sets the weight for a specific node (for WeightedRoundRobin strategy).

```go
func (p *Pool) SetNodeWeight(url string, weight int) error
```

**Parameters:**
- `url` - Node URL to set weight for
- `weight` - Weight value (higher = more requests)

**Returns:**
- `error` - Error if setting weight fails

**Example:**
```go
// Set different weights for nodes with different capacities
pool.SetNodeWeight("amqp://node1:5672/", 3)  // High capacity
pool.SetNodeWeight("amqp://node2:5672/", 2)  // Medium capacity
pool.SetNodeWeight("amqp://node3:5672/", 1)  // Lower capacity
```

### GetHealthyNodeCount

Gets the count of currently healthy nodes.

```go
func (p *Pool) GetHealthyNodeCount() int
```

**Returns:**
- `int` - Number of healthy nodes

**Example:**
```go
healthyCount := pool.GetHealthyNodeCount()
if healthyCount == 0 {
    log.Fatal("No healthy nodes available")
}
log.Printf("Available nodes: %d", healthyCount)
```

## Configuration

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

**Fields:**
- `URLs` - List of RabbitMQ node URLs (required)
- `ReconnectInterval` - Time between reconnection attempts (default: `5s`)
- `MaxReconnectAttempt` - Maximum number of reconnection attempts (default: `10`)
- `HealthCheckInterval` - Interval between health checks (default: `30s`)
- `LoadBalanceStrategy` - Load balancing strategy (default: `RoundRobin`)
- `DebugLog` - Enable/disable debug logging (default: `false`)
- `Logger` - Custom logger implementation (default: `DefaultLogger`)
- `TLSConfig` - TLS/SSL configuration (optional)

### LoadBalanceStrategy

Available load balancing strategies.

```go
const (
    RoundRobin         LoadBalanceStrategy = iota // Distribute requests evenly
    Random                                       // Randomly select nodes
    LeastUsed                                   // Select least used node
    WeightedRoundRobin                          // Distribute based on weights
)
```

## Data Structures

### PoolStats

Pool statistics structure.

```go
type PoolStats struct {
    TotalNodes     int           // Total number of nodes
    HealthyNodes   int           // Number of healthy nodes
    UnhealthyNodes int           // Number of unhealthy nodes
    LastHealthCheck time.Time    // Last health check timestamp
    LoadBalanceStrategy LoadBalanceStrategy // Current strategy
}
```

### NodeConnection

Internal node connection structure.

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

## Load Balancing Strategies

### Round Robin (Default)

Distributes requests evenly across all healthy nodes in sequence.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.RoundRobin,
}
```

**Advantages:**
- Even distribution
- Simple and predictable
- Good for most use cases

### Random

Randomly selects a healthy node for each request.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.Random,
}
```

**Advantages:**
- Good for high-throughput scenarios
- Minimizes predictable patterns
- Reduces local overload

### Least Used

Selects the node with the lowest usage count.

```go
config := bunnyhop.PoolConfig{
    LoadBalanceStrategy: bunnyhop.LeastUsed,
}
```

**Advantages:**
- Optimizes resource usage
- Automatic load balancing
- Good for resource-intensive operations

### Weighted Round Robin

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

## Health Monitoring

### Automatic Health Checks

The pool automatically monitors node health:

- **Connection Monitoring**: Watches for connection drops
- **Health Checks**: Periodic health checks on all nodes
- **Auto-Reconnection**: Automatically reconnects to failed nodes
- **Load Balancing**: Routes requests only to healthy nodes

### Health Check Configuration

```go
config := bunnyhop.PoolConfig{
    HealthCheckInterval: 15 * time.Second, // More frequent checks
    ReconnectInterval:   3 * time.Second,  // Faster reconnection
    MaxReconnectAttempt: 15,               // More attempts
}
```

## Error Handling

### Connection Failures

The pool automatically handles:

- **Network failures**: Automatic reconnection
- **Node failures**: Failover to healthy nodes
- **Authentication errors**: Retry with exponential backoff
- **TLS errors**: Fallback to non-TLS if configured

### Error Scenarios

```go
client, err := pool.GetClient()
if err != nil {
    switch {
    case errors.Is(err, ErrNoHealthyNodes):
        log.Fatal("No healthy nodes available")
    case errors.Is(err, ErrPoolNotStarted):
        log.Fatal("Pool not started")
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## Best Practices

1. **Always close the pool** using `defer pool.Close()`
2. **Check pool health** before operations using `GetHealthyNodeCount()`
3. **Monitor statistics** using `GetStats()` for production monitoring
4. **Use appropriate strategy** based on your workload requirements
5. **Set node weights** for WeightedRoundRobin to optimize distribution
6. **Handle errors gracefully** and implement retry logic for critical operations

## Examples

### Basic Pool Usage

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
        log.Fatalf("Failed to start pool: %v", err)
    }
    
    // Wait for pool to be ready
    time.Sleep(2 * time.Second)
    
    // Get client using load balancing
    client, err := pool.GetClient()
    if err != nil {
        log.Fatalf("Failed to get client: %v", err)
    }
    
    // Use the client
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Printf("Failed to declare queue: %v", err)
        return
    }
    
    log.Printf("Queue declared: %s", queue.Name)
    
    // Get pool statistics
    stats := pool.GetStats()
    log.Printf("Pool stats: %+v", stats)
}
```

### Production Pool Configuration

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
    
    // Set weights for different node capacities
    pool.SetNodeWeight("amqp://prod1:5672/", 3)  // High capacity
    pool.SetNodeWeight("amqp://prod2:5672/", 2)  // Medium capacity
    pool.SetNodeWeight("amqp://prod3:5672/", 1)  // Lower capacity
    
    return pool
}
```

## Next Steps

- Read [Client API Reference](client.md) for client operations
- Check [Configuration Guide](../configuration.md) for detailed options
- Review [Production Deployment](../production.md) for best practices
- Explore [Examples](../examples/) for usage patterns
