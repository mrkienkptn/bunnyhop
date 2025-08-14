# Basic Examples

This section provides basic examples of common RabbitMQ operations using BunnyHop.

## Simple Message Publishing

### Basic Publisher

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
    // Create client configuration
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    // Create new client
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Connect to RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Declare exchange
    err := client.DeclareExchange("test_exchange", "direct", true, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to declare exchange: %v", err)
    }
    
    // Declare queue
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }
    
    // Bind queue to exchange
    err = client.QueueBind("test_queue", "test_key", "test_exchange", false, nil)
    if err != nil {
        log.Fatalf("Failed to bind queue: %v", err)
    }
    
    // Publish message
    message := amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte("Hello, RabbitMQ!"),
        Timestamp:   time.Now(),
    }
    
    err = client.PublishMessage("test_exchange", "test_key", false, false, message)
    if err != nil {
        log.Fatalf("Failed to publish message: %v", err)
    }
    
    log.Printf("Message published to queue: %s", queue.Name)
}
```

## Message Consumer

### Basic Consumer

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
    // Create client configuration
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    // Create new client
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Connect to RabbitMQ
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    
    // Get channel for consumer operations
    channel, err := client.GetChannel()
    if err != nil {
        log.Fatalf("Failed to get channel: %v", err)
    }
    
    // Declare queue (same as publisher)
    queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }
    
    // Set QoS for fair distribution
    err = channel.Qos(1, 0, false)
    if err != nil {
        log.Fatalf("Failed to set QoS: %v", err)
    }
    
    // Start consuming messages
    msgs, err := channel.Consume(
        queue.Name, // queue
        "",         // consumer
        false,      // auto-ack
        false,      // exclusive
        false,      // no-local
        false,      // no-wait
        nil,        // args
    )
    if err != nil {
        log.Fatalf("Failed to start consuming: %v", err)
    }
    
    log.Printf("Started consuming from queue: %s", queue.Name)
    
    // Process messages
    for msg := range msgs {
        log.Printf("Received message: %s", string(msg.Body))
        
        // Process the message here
        time.Sleep(100 * time.Millisecond) // Simulate processing
        
        // Acknowledge the message
        msg.Ack(false)
    }
}
```

## Using Connection Pool

### Pool-based Publisher

```go
package main

import (
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
            "amqp://guest:guest@localhost:5673/",
        },
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        HealthCheckInterval: 30 * time.Second,
        LoadBalanceStrategy: bunnyhop.RoundRobin,
        DebugLog:            true,
    }
    
    // Create new pool
    pool := bunnyhop.NewPool(config)
    defer pool.Close()
    
    // Start the pool
    if err := pool.Start(); err != nil {
        log.Fatalf("Failed to start pool: %v", err)
    }
    
    // Wait for pool to be ready
    time.Sleep(2 * time.Second)
    
    // Get client from pool
    client, err := pool.GetClient()
    if err != nil {
        log.Fatalf("Failed to get client: %v", err)
    }
    
    // Use the client for operations
    queue, err := client.DeclareQueue("pool_queue", true, false, false, nil)
    if err != nil {
        log.Fatalf("Failed to declare queue: %v", err)
    }
    
    // Publish multiple messages
    for i := 0; i < 10; i++ {
        message := amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(fmt.Sprintf("Message %d from pool", i)),
        }
        
        err = client.PublishMessage("", queue.Name, false, false, message)
        if err != nil {
            log.Printf("Failed to publish message %d: %v", i, err)
        } else {
            log.Printf("Message %d published successfully", i)
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // Get pool statistics
    stats := pool.GetStats()
    log.Printf("Pool stats: %+v", stats)
}
```

## Error Handling

### Robust Error Handling

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    config := bunnyhop.Config{
        URLs:                []string{"amqp://guest:guest@localhost:5672/"},
        ReconnectInterval:   5 * time.Second,
        MaxReconnectAttempt: 10,
        DebugLog:            true,
    }
    
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Retry connection with exponential backoff
    maxRetries := 5
    for attempt := 0; attempt < maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        
        err := client.Connect(ctx)
        cancel()
        
        if err == nil {
            log.Println("Successfully connected to RabbitMQ")
            break
        }
        
        log.Printf("Connection attempt %d failed: %v", attempt+1, err)
        
        if attempt < maxRetries-1 {
            backoff := time.Duration(attempt+1) * time.Second
            log.Printf("Retrying in %v...", backoff)
            time.Sleep(backoff)
        } else {
            log.Fatalf("Failed to connect after %d attempts", maxRetries)
        }
    }
    
    // Check connection health
    if !client.IsConnected() {
        log.Fatal("Client is not connected")
    }
    
    log.Println("Ready to perform operations")
}
```

## Configuration Examples

### Environment-based Configuration

```go
package main

import (
    "log"
    "os"
    "strconv"
    "time"
    
    "github.com/vanduc0209/bunnyhop"
)

func getConfig() bunnyhop.Config {
    // Get configuration from environment variables
    urls := []string{"amqp://localhost:5672"} // Default
    if envURLs := os.Getenv("RABBITMQ_URLS"); envURLs != "" {
        urls = strings.Split(envURLs, ",")
    }
    
    reconnectInterval := 5 * time.Second // Default
    if envInterval := os.Getenv("RABBITMQ_RECONNECT_INTERVAL"); envInterval != "" {
        if interval, err := time.ParseDuration(envInterval); err == nil {
            reconnectInterval = interval
        }
    }
    
    maxAttempts := 10 // Default
    if envAttempts := os.Getenv("RABBITMQ_MAX_RECONNECT_ATTEMPTS"); envAttempts != "" {
        if attempts, err := strconv.Atoi(envAttempts); err == nil {
            maxAttempts = attempts
        }
    }
    
    debugLog := false // Default
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

func main() {
    config := getConfig()
    log.Printf("Using configuration: %+v", config)
    
    client := bunnyhop.NewClient(config)
    defer client.Close()
    
    // Use the client...
}
```

## Running the Examples

1. **Start RabbitMQ** (using Docker):
   ```bash
   make docker-single
   ```

2. **Run the publisher**:
   ```bash
   go run publisher.go
   ```

3. **Run the consumer** (in another terminal):
   ```bash
   go run consumer.go
   ```

## Next Steps

- Explore [Advanced Examples](advanced.md) for complex scenarios
- Check [Integration Examples](integration.md) for real-world use cases
- Review [API Reference](../api/) for complete method documentation
- Read [Production Deployment](../production.md) for best practices 