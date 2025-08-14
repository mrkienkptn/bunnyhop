# Quick Start Guide

Get up and running with BunnyHop in minutes! This guide will show you how to create your first RabbitMQ connection and start sending messages.

## Prerequisites

- Go 1.24+ installed
- RabbitMQ running (see [Installation Guide](installation.md))
- Basic understanding of Go

## Your First Connection

### 1. Create a Simple Client

Create a new Go file and add this code:

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
    
    log.Println("Successfully connected to RabbitMQ!")
    
    // Don't forget to close the connection
    defer client.Close()
}
```

### 2. Run Your Code

```bash
go run main.go
```

You should see: `Successfully connected to RabbitMQ!`

## Basic Operations

### Declare a Queue

```go
// Declare a durable queue
queue, err := client.DeclareQueue("my_queue", true, false, false, nil)
if err != nil {
    log.Printf("Failed to declare queue: %v", err)
} else {
    log.Printf("Queue declared: %s", queue.Name)
}
```

### Declare an Exchange

```go
// Declare a direct exchange
err = client.DeclareExchange("my_exchange", "direct", true, false, false, nil)
if err != nil {
    log.Printf("Failed to declare exchange: %v", err)
} else {
    log.Println("Exchange declared successfully")
}
```

### Bind Queue to Exchange

```go
// Bind queue to exchange with routing key
err = client.QueueBind("my_queue", "my_key", "my_exchange", false, nil)
if err != nil {
    log.Printf("Failed to bind queue: %v", err)
} else {
    log.Println("Queue bound successfully")
}
```

### Publish a Message

```go
// Create a message
message := amqp.Publishing{
    ContentType: "text/plain",
    Body:        []byte("Hello, RabbitMQ!"),
}

// Publish to exchange
err = client.PublishMessage("my_exchange", "my_key", false, false, message)
if err != nil {
    log.Printf("Failed to publish message: %v", err)
} else {
    log.Println("Message published successfully")
}
```

## Using Connection Pool

For production applications, use the connection pool for better reliability:

### 1. Create Pool Configuration

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

### 2. Create and Start Pool

```go
// Create new pool
pool := bunnyhop.NewPool(config)

// Start the pool
if err := pool.Start(); err != nil {
    log.Fatalf("Failed to start pool: %v", err)
}

// Don't forget to close the pool
defer pool.Close()
```

### 3. Get Client from Pool

```go
// Get client using load balancing strategy
client, err := pool.GetClient()
if err != nil {
    log.Printf("Failed to get client: %v", err)
    return
}

// Use the client as before
queue, err := client.DeclareQueue("pool_queue", true, false, false, nil)
if err != nil {
    log.Printf("Failed to declare queue: %v", err)
} else {
    log.Printf("Queue declared: %s", queue.Name)
}
```

## Complete Example

Here's a complete working example:

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
    // Create pool configuration
    config := bunnyhop.PoolConfig{
        URLs: []string{
            "amqp://guest:guest@localhost:5672/",
        },
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        HealthCheckInterval: 30 * time.Second,
        LoadBalanceStrategy: bunnyhop.RoundRobin,
        DebugLog:            true,
    }
    
    // Create and start pool
    pool := bunnyhop.NewPool(config)
    if err := pool.Start(); err != nil {
        log.Fatalf("Failed to start pool: %v", err)
    }
    defer pool.Close()
    
    // Wait for pool to be ready
    time.Sleep(2 * time.Second)
    
    // Get client from pool
    client, err := pool.GetClient()
    if err != nil {
        log.Printf("Failed to get client: %v", err)
        return
    }
    
    // Declare queue
    queue, err := client.DeclareQueue("hello", true, false, false, nil)
    if err != nil {
        log.Printf("Failed to declare queue: %v", err)
        return
    }
    log.Printf("Queue declared: %s", queue.Name)
    
    // Declare exchange
    err = client.DeclareExchange("hello_exchange", "direct", true, false, false, nil)
    if err != nil {
        log.Printf("Failed to declare exchange: %v", err)
        return
    }
    
    // Bind queue to exchange
    err = client.QueueBind("hello", "hello_key", "hello_exchange", false, nil)
    if err != nil {
        log.Printf("Failed to bind queue: %v", err)
        return
    }
    
    // Publish message
    message := amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte("Hello, BunnyHop!"),
    }
    
    err = client.PublishMessage("hello_exchange", "hello_key", false, false, message)
    if err != nil {
        log.Printf("Failed to publish message: %v", err)
        return
    }
    
    log.Println("Message published successfully!")
    
    // Get pool statistics
    stats := pool.GetStats()
    log.Printf("Pool stats: %+v", stats)
}
```

## What's Next?

Now that you have the basics:

1. **Explore Examples**: Check out [Basic Examples](examples/basic.md) for more patterns
2. **Learn Configuration**: Read [Configuration Guide](configuration.md) for advanced options
3. **Understand Architecture**: See [Architecture Overview](architecture.md) for how it works
4. **Production Ready**: Review [Production Deployment](production.md) for best practices

## Need Help?

- Check the [Troubleshooting Guide](troubleshooting.md)
- Review [API Reference](api/) for detailed method documentation
- Open an issue on [GitHub](https://github.com/vanduc0209/bunnyhop/issues) 