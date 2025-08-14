# Architecture Overview

This document provides a comprehensive overview of BunnyHop's architecture, design principles, and internal components.

## Overview

BunnyHop is a Go library built on top of `amqp091-go` that provides high-level abstractions for RabbitMQ operations. It's designed with the following principles:

- **Simplicity**: Easy-to-use API that abstracts away RabbitMQ complexity
- **Reliability**: Built-in connection management, reconnection, and failover
- **Performance**: Efficient connection pooling and load balancing
- **Extensibility**: Pluggable logging, health checking, and monitoring

## High-Level Architecture

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

## Core Components

### 1. Client

The `Client` type provides a single connection to RabbitMQ with automatic reconnection capabilities.

**Key Features:**
- Single connection management
- Automatic reconnection with configurable retry logic
- Connection health monitoring
- Thread-safe operations

**Architecture:**
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

**Internal Structure:**
```go
type Client struct {
    config     Config
    conn       *amqp.Connection
    channel    *amqp.Channel
    logger     Logger
    mu         sync.RWMutex
    isConnected bool
    // ... other fields
}
```

### 2. Pool

The `Pool` type manages multiple RabbitMQ connections with load balancing and failover capabilities.

**Key Features:**
- Multiple node support
- Load balancing strategies
- Automatic failover
- Health monitoring
- Connection pooling

**Architecture:**
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

**Internal Structure:**
```go
type Pool struct {
    config     PoolConfig
    nodes      map[string]*NodeConnection
    strategy   LoadBalanceStrategy
    logger     Logger
    mu         sync.RWMutex
    // ... other fields
}
```

### 3. NodeConnection

Internal structure for managing individual node connections within a pool.

**Architecture:**
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

## Load Balancing Strategies

### 1. Round Robin (Default)

Distributes requests evenly across all healthy nodes in sequence.

**Algorithm:**
```go
func (p *Pool) getNextRoundRobin() *NodeConnection {
    p.currentIndex = (p.currentIndex + 1) % len(p.healthyNodes)
    return p.healthyNodes[p.currentIndex]
}
```

**Advantages:**
- Predictable distribution
- Simple implementation
- Good for most use cases

### 2. Random

Randomly selects a healthy node for each request.

**Algorithm:**
```go
func (p *Pool) getNextRandom() *NodeConnection {
    if len(p.healthyNodes) == 0 {
        return nil
    }
    randomIndex := rand.Intn(len(p.healthyNodes))
    return p.healthyNodes[randomIndex]
}
```

**Advantages:**
- Reduces predictable patterns
- Good for high-throughput scenarios
- Minimizes local overload

### 3. Least Used

Selects the node with the lowest usage count.

**Algorithm:**
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

**Advantages:**
- Optimizes resource usage
- Automatic load balancing
- Good for resource-intensive operations

### 4. Weighted Round Robin

Distributes requests based on node weights.

**Algorithm:**
```go
func (p *Pool) getNextWeightedRoundRobin() *NodeConnection {
    // Complex algorithm that considers weights
    // and distributes requests proportionally
}
```

**Advantages:**
- Flexible and customizable
- Suitable for nodes with different capacities
- Good control over distribution

## Health Monitoring

### Health Check Process

The pool continuously monitors node health through several mechanisms:

1. **Connection Monitoring**: Watches for connection drops
2. **Periodic Health Checks**: Regular health check intervals
3. **Error Tracking**: Monitors failed operations
4. **Auto-Reconnection**: Attempts to reconnect failed nodes

**Health Check Flow:**
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

### Health Check Implementation

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

## Connection Management

### Connection Lifecycle

1. **Initialization**: Create connection with configuration
2. **Establishment**: Connect to RabbitMQ server
3. **Monitoring**: Watch connection health
4. **Reconnection**: Automatically reconnect on failure
5. **Cleanup**: Properly close connections

**Connection States:**
```
Disconnected ──▶ Connecting ──▶ Connected
     ▲                              │
     │                              ▼
     └── Reconnecting ◀─── Failed ◀──┘
```

### Reconnection Logic

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

## Error Handling

### Error Types

BunnyHop defines several custom error types for better error handling:

```go
var (
    ErrNoHealthyNodes    = errors.New("no healthy nodes available")
    ErrPoolNotStarted    = errors.New("pool not started")
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrConnectionFailed  = errors.New("connection failed")
    ErrReconnectFailed   = errors.New("reconnection failed")
)
```

### Error Handling Strategy

1. **Graceful Degradation**: Continue operation with available nodes
2. **Retry Logic**: Automatic retry for transient failures
3. **Circuit Breaker**: Prevent cascading failures
4. **Error Propagation**: Clear error messages for debugging

## Thread Safety

### Concurrency Model

BunnyHop uses a combination of mutexes and channels to ensure thread safety:

```go
type Pool struct {
    mu    sync.RWMutex
    nodes map[string]*NodeConnection
    // ... other fields
}

func (p *Pool) GetClient() (*Client, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    // Thread-safe operations
}
```

### Locking Strategy

- **Read-Write Mutexes**: Allow concurrent reads, exclusive writes
- **Fine-grained Locking**: Lock only when necessary
- **Deadlock Prevention**: Consistent lock ordering

## Performance Considerations

### Connection Pooling

- **Reuse Connections**: Minimize connection overhead
- **Channel Pooling**: Efficient channel management
- **Lazy Initialization**: Create resources on demand

### Memory Management

- **Object Reuse**: Minimize allocations
- **Buffer Pooling**: Reuse buffers for messages
- **Garbage Collection**: Minimize GC pressure

### Optimization Techniques

1. **Connection Reuse**: Keep connections alive
2. **Batch Operations**: Group multiple operations
3. **Async Processing**: Non-blocking operations
4. **Resource Limits**: Prevent resource exhaustion

## Monitoring and Observability

### Metrics Collection

BunnyHop provides built-in metrics for monitoring:

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

Structured logging with configurable levels:

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

### Health Endpoints

Built-in health check endpoints for monitoring systems:

```go
func (p *Pool) HealthCheck() HealthStatus {
    return HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Details:   p.GetStats(),
    }
}
```

## Security

### Authentication

- **Username/Password**: Standard AMQP authentication
- **TLS/SSL**: Encrypted connections
- **Virtual Hosts**: Resource isolation

### Network Security

- **Connection Encryption**: TLS support
- **Access Control**: RabbitMQ permissions
- **Network Isolation**: Firewall considerations

## Deployment Considerations

### Single Node Deployment

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

### Cluster Deployment

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

### Load Balancer Integration

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

## Future Enhancements

### Planned Features

1. **Circuit Breaker Pattern**: Advanced failure handling
2. **Rate Limiting**: Prevent overwhelming RabbitMQ
3. **Metrics Export**: Prometheus/OpenTelemetry integration
4. **Distributed Tracing**: Request tracing across nodes
5. **Plugin System**: Extensible architecture

### Architecture Evolution

- **Microservices Support**: Better integration patterns
- **Event Sourcing**: Advanced event handling
- **CQRS Support**: Command/Query separation
- **Saga Pattern**: Distributed transaction support

## Best Practices

### 1. Connection Management

- Always close connections properly
- Use connection pooling for high-throughput applications
- Monitor connection health regularly

### 2. Error Handling

- Implement proper error handling and retry logic
- Use custom error types for better debugging
- Log errors with sufficient context

### 3. Performance

- Choose appropriate load balancing strategy
- Monitor performance metrics
- Optimize based on workload patterns

### 4. Security

- Use TLS for production deployments
- Implement proper authentication
- Follow security best practices

## Next Steps

- Read [Configuration Guide](configuration.md) for setup details
- Check [API Reference](api/) for detailed API documentation
- Review [Examples](examples/) for usage patterns
- Explore [Production Guide](production.md) for deployment best practices
