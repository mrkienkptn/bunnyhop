# BunnyHop

[![Go Report Card](https://goreportcard.com/badge/github.com/vanduc0209/bunnyhop)](https://goreportcard.com/report/github.com/vanduc0209/go-httpx)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vanduc0209/bunnyhop)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A Go library built on top of [amqp091-go](https://github.com/rabbitmq/amqp091-go) that provides seamless RabbitMQ cluster connections with automatic retry, failover, and multi-node connection management. Designed for high-availability and load-balanced messaging in production environments.

## Features

- 🚀 **Cluster Support**: Connect to multiple RabbitMQ nodes with automatic failover
- 🔄 **Auto-Reconnection**: Automatic reconnection with configurable retry strategies
- ⚖️ **Load Balancing**: Multiple load balancing strategies (Round Robin, Random, Least Used, Weighted Round Robin)
- 🏥 **Health Monitoring**: Built-in health checks and connection monitoring
- 📊 **Metrics & Statistics**: Comprehensive connection and usage statistics
- 🐛 **Debug Logging**: Configurable debug logging for development and troubleshooting
- 🎯 **Production Ready**: Designed for high-availability production environments

## Installation

```bash
go get github.com/vanduc0209/bunnyhop
```

## Quick Start

### Single Node Connection

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/vanduc0209/bunnyhop"
)

func main() {
    // Create client configuration
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }

    // Create new client
    client := bunnyhop.NewClient(config)

    // Connect to RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // Use the client
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Printf("Failed to declare queue: %v", err)
    } else {
        log.Printf("Queue declared: %s", queue.Name)
    }

    // Close connection
    client.Close()
}
```

### Cluster Connection with Pool

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/vanduc0209/bunnyhop"
)

func main() {
    // Create pool configuration
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

    // Create new pool
    pool := bunnyhop.NewPool(config)

    // Start the pool
    if err := pool.Start(); err != nil {
        log.Fatalf("Failed to start pool: %v", err)
    }

    // Get client from pool
    client, err := pool.GetClient()
    if err != nil {
        log.Printf("Failed to get client: %v", err)
        return
    }

    // Use the client
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Printf("Failed to declare queue: %v", err)
    } else {
        log.Printf("Queue declared: %s", queue.Name)
    }

    // Get pool statistics
    stats := pool.GetStats()
    log.Printf("Pool stats: %+v", stats)

    // Close the pool
    pool.Close()
}
```

## Docker Setup

### Single Node RabbitMQ

Quick start with a single RabbitMQ node:

```bash
# Using Makefile
make docker-single

# Or using script
chmod +x scripts/start-single.sh
./scripts/start-single.sh

# Or manually
docker-compose -f docker-compose.single.yml up -d
```

**Access:**
- AMQP: `localhost:5672`
- Management UI: `http://localhost:15672` (guest/guest)

### 3-Node Cluster

Start a full RabbitMQ cluster with HAProxy load balancer:

```bash
# Using Makefile
make docker-cluster

# Or using script
chmod +x scripts/start-cluster.sh
./scripts/start-cluster.sh

# Or manually
docker-compose -f docker-compose.cluster.yml up -d
```

**Access:**
- **Individual Nodes:**
  - Node 1: AMQP `localhost:5672`, Management `http://localhost:15672`
  - Node 2: AMQP `localhost:5673`, Management `http://localhost:15673`
  - Node 3: AMQP `localhost:5674`, Management `http://localhost:15674`
- **Load Balancer (HAProxy):**
  - AMQP: `localhost:5670`
  - Management UI: `http://localhost:15670`
  - Stats: `http://localhost:8404` (admin/admin123)

### Docker Commands

```bash
# View running containers
make docker-ps

# View logs
make docker-logs-single    # Single node
make docker-logs-cluster   # Cluster

# Stop containers
make docker-stop

# Clean up (removes volumes)
make docker-clean
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run specific test file
go test -v ./client_test.go
go test -v ./pool_test.go
go test -v ./logger_test.go

# Run tests with coverage
make coverage
```

### Integration Tests

Integration tests require RabbitMQ to be running:

```bash
# Start RabbitMQ first
make docker-single

# Run integration tests
go test -v -tags=integration ./...

# Run specific integration test
go test -v -run TestClientIntegration_SingleNode
go test -v -run TestPoolIntegration_Cluster
```

### Benchmarks

```bash
# Run benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkClient_Connect
go test -bench=BenchmarkPool_GetClient
```

## Configuration

### Client Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `URLs` | `[]string` | `["amqp://localhost:5672"]` | List of RabbitMQ connection URLs |
| `ReconnectInterval` | `time.Duration` | `5s` | Time between reconnection attempts |
| `MaxReconnectAttempt` | `int` | `10` | Maximum number of reconnection attempts |
| `DebugLog` | `bool` | `false` | Enable/disable debug logging |
| `Logger` | `Logger` | `DefaultLogger` | Custom logger implementation |

### Pool Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `URLs` | `[]string` | `["amqp://localhost:5672"]` | List of RabbitMQ node URLs |
| `ReconnectInterval` | `time.Duration` | `5s` | Time between reconnection attempts |
| `MaxReconnectAttempt` | `int` | `10` | Maximum number of reconnection attempts |
| `HealthCheckInterval` | `time.Duration` | `30s` | Interval between health checks |
| `LoadBalanceStrategy` | `LoadBalanceStrategy` | `RoundRobin` | Load balancing strategy |
| `DebugLog` | `bool` | `false` | Enable/disable debug logging |
| `Logger` | `Logger` | `DefaultLogger` | Custom logger implementation |

## Load Balancing Strategies

### Round Robin
Distributes requests evenly across all healthy nodes in sequence.

### Random
Randomly selects a healthy node for each request.

### Least Used
Selects the node with the lowest usage count.

### Weighted Round Robin
Distributes requests based on node weights, with higher weights receiving more requests.

## API Reference

### Client Methods

- `Connect(ctx context.Context) error` - Establish connection to RabbitMQ
- `IsConnected() bool` - Check if client is connected
- `GetChannel() (*amqp.Channel, error)` - Get current AMQP channel
- `GetConnection() (*amqp.Connection, error)` - Get current AMQP connection
- `Close() error` - Close connection
- `PublishMessage(exchange, routingKey string, mandatory, immediate bool, msg amqp.Publishing) error` - Publish message
- `DeclareQueue(name string, durable, autoDelete, exclusive bool, args amqp.Table) (amqp.Queue, error)` - Declare queue
- `DeclareExchange(name, kind string, durable, autoDelete, internal bool, args amqp.Table) error` - Declare exchange
- `QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error` - Bind queue to exchange

### Pool Methods

- `Start() error` - Start the pool and establish connections
- `GetClient() (*Client, error)` - Get a client using load balancing strategy
- `GetStats() PoolStats` - Get pool statistics
- `Close() error` - Close the pool and all connections
- `SetNodeWeight(url string, weight int) error` - Set weight for a specific node
- `GetHealthyNodeCount() int` - Get count of healthy nodes

## Logging

BunnyHop provides comprehensive logging with different levels:

- **DEBUG**: Detailed connection and operation information
- **INFO**: General operational information
- **WARN**: Warning messages for non-critical issues
- **ERROR**: Error messages for connection and operation failures

You can implement a custom logger by implementing the `Logger` interface:

```go
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}
```

## Health Monitoring

The pool automatically monitors the health of all nodes:

- **Connection Monitoring**: Watches for connection drops
- **Health Checks**: Periodic health checks on all nodes
- **Auto-Reconnection**: Automatically reconnects to failed nodes
- **Load Balancing**: Routes requests only to healthy nodes

## Production Considerations

- **Connection Limits**: Monitor connection counts and adjust pool sizes
- **Error Handling**: Implement proper error handling for connection failures
- **Monitoring**: Use the built-in statistics for monitoring and alerting
- **Load Balancing**: Choose appropriate load balancing strategy for your use case
- **Logging**: Configure appropriate log levels for production environments

## Development

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- Make (optional, for convenience)

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/vanduc0209/bunnyhop.git
cd bunnyhop

# Install dependencies
make deps

# Run tests
make test

# Start RabbitMQ for testing
make docker-single
```

### Project Structure

```
bunnyhop/
├── client.go              # RabbitMQ client implementation
├── pool.go                # Connection pool implementation
├── types.go               # Type definitions and interfaces
├── logger.go              # Logger interface and default implementation
├── utils.go               # Utility functions
├── *_test.go              # Unit tests
├── integration_test.go    # Integration tests
├── example/               # Usage examples
├── scripts/               # Shell scripts for Docker management
├── docker-compose.*.yml   # Docker Compose configurations
├── haproxy.cfg            # HAProxy configuration for cluster
├── Dockerfile.test        # Dockerfile for test client
├── Makefile               # Build and management commands
└── README.md              # This file
```

## Examples

See the `example/` directory for complete working examples.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
