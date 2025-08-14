# API Reference

This section provides comprehensive API documentation for BunnyHop.

## Available APIs

### Core APIs

- **[Client API](client.md)** - Complete reference for the `Client` type
- **[Pool API](pool.md)** - Complete reference for the `Pool` type  
- **[Types & Interfaces](types.md)** - All types, interfaces, and data structures

### Quick Reference

| Component | Purpose | Key Methods |
|-----------|---------|-------------|
| **Client** | Single RabbitMQ connection | `Connect()`, `DeclareQueue()`, `PublishMessage()`, `Consume()` |
| **Pool** | Multiple connections with load balancing | `Start()`, `GetClient()`, `GetStats()`, `Close()` |

### Configuration

| Type | Description | Key Fields |
|------|-------------|------------|
| **Config** | Client configuration | `URLs`, `ReconnectInterval`, `MaxReconnectAttempt`, `DebugLog` |
| **PoolConfig** | Pool configuration | `URLs`, `LoadBalanceStrategy`, `HealthCheckInterval`, `Logger` |

### Load Balancing Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| **RoundRobin** | Distribute evenly in sequence | General purpose, predictable |
| **Random** | Random selection | High throughput, load distribution |
| **LeastUsed** | Select least used node | Resource optimization |
| **WeightedRoundRobin** | Distribute based on weights | Different node capacities |

## Getting Started

1. **Choose your approach**:
   - Use `Client` for simple, single-connection scenarios
   - Use `Pool` for high-availability, multi-node scenarios

2. **Configure your setup**:
   - Set connection URLs
   - Choose load balancing strategy (for pools)
   - Configure reconnection settings

3. **Implement error handling**:
   - Handle connection failures
   - Implement retry logic
   - Monitor health status

## Examples

### Basic Client Usage

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

### Pool with Load Balancing

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

## Error Handling

BunnyHop provides custom error types for better error handling:

```go
var (
    ErrNoHealthyNodes    = errors.New("no healthy nodes available")
    ErrPoolNotStarted    = errors.New("pool not started")
    ErrInvalidConfig     = errors.New("invalid configuration")
    ErrConnectionFailed  = errors.New("connection failed")
    ErrReconnectFailed   = errors.New("reconnection failed")
)
```

## Thread Safety

- **Client**: Thread-safe for concurrent operations
- **Pool**: Thread-safe with read-write mutexes
- **All operations**: Safe for concurrent access

## Performance Considerations

- **Connection reuse**: Minimize connection overhead
- **Channel pooling**: Efficient channel management
- **Load balancing**: Distribute load across nodes
- **Health monitoring**: Automatic failover

## Next Steps

- Read individual API documentation for detailed information
- Check [Configuration Guide](../configuration.md) for setup details
- Review [Examples](../examples/) for usage patterns
- Explore [Production Guide](../production.md) for deployment best practices
