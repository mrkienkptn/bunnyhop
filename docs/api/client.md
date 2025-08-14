# Client API Reference

The `Client` type provides a single connection to RabbitMQ with automatic reconnection capabilities.

## Overview

```go
type Client struct {
    // Private fields for internal use
}
```

## Constructor

### NewClient

Creates a new RabbitMQ client with the specified configuration.

```go
func NewClient(config Config) *Client
```

**Parameters:**
- `config` - Client configuration

**Returns:**
- `*Client` - New client instance

**Example:**
```go
config := bunnyhop.Config{
    URLs:                []string{"amqp://guest:guest@localhost:5672/"},
    ReconnectInterval:   5 * time.Second,
    MaxReconnectAttempt: 10,
    DebugLog:            true,
}

client := bunnyhop.NewClient(config)
```

## Methods

### Connect

Establishes a connection to RabbitMQ.

```go
func (c *Client) Connect(ctx context.Context) error
```

**Parameters:**
- `ctx` - Context for the connection operation

**Returns:**
- `error` - Connection error if any

**Example:**
```go
ctx := context.Background()
if err := client.Connect(ctx); err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### IsConnected

Checks if the client is currently connected to RabbitMQ.

```go
func (c *Client) IsConnected() bool
```

**Returns:**
- `bool` - True if connected, false otherwise

**Example:**
```go
if client.IsConnected() {
    log.Println("Client is connected")
} else {
    log.Println("Client is disconnected")
}
```

### GetChannel

Retrieves the current AMQP channel.

```go
func (c *Client) GetChannel() (*amqp.Channel, error)
```

**Returns:**
- `*amqp.Channel` - Current AMQP channel
- `error` - Error if client is not connected

**Example:**
```go
channel, err := client.GetChannel()
if err != nil {
    log.Printf("Failed to get channel: %v", err)
    return
}

// Use the channel for AMQP operations
```

### GetConnection

Retrieves the current AMQP connection.

```go
func (c *Client) GetConnection() (*amqp.Connection, error)
```

**Returns:**
- `*amqp.Connection` - Current AMQP connection
- `error` - Error if client is not connected

**Example:**
```go
connection, err := client.GetConnection()
if err != nil {
    log.Printf("Failed to get connection: %v", err)
    return
}

// Use the connection for advanced operations
```

### Close

Closes the client connection and cleans up resources.

```go
func (c *Client) Close() error
```

**Returns:**
- `error` - Any errors encountered during cleanup

**Example:**
```go
defer func() {
    if err := client.Close(); err != nil {
        log.Printf("Error closing client: %v", err)
    }
}()
```

## RabbitMQ Operations

### PublishMessage

Publishes a message to an exchange.

```go
func (c *Client) PublishMessage(
    exchange, routingKey string,
    mandatory, immediate bool,
    msg amqp.Publishing,
) error
```

**Parameters:**
- `exchange` - Exchange name
- `routingKey` - Routing key for the message
- `mandatory` - Whether the message is mandatory
- `immediate` - Whether the message is immediate
- `msg` - Message to publish

**Returns:**
- `error` - Publishing error if any

**Example:**
```go
message := amqp.Publishing{
    ContentType: "text/plain",
    Body:        []byte("Hello, RabbitMQ!"),
}

err := client.PublishMessage("my_exchange", "my_key", false, false, message)
if err != nil {
    log.Printf("Failed to publish message: %v", err)
}
```

### DeclareQueue

Declares a queue on RabbitMQ.

```go
func (c *Client) DeclareQueue(
    name string,
    durable, autoDelete, exclusive bool,
    args amqp.Table,
) (amqp.Queue, error)
```

**Parameters:**
- `name` - Queue name
- `durable` - Whether the queue survives broker restart
- `autoDelete` - Whether the queue is deleted when last consumer unsubscribes
- `exclusive` - Whether the queue is exclusive to this connection
- `args` - Additional queue arguments

**Returns:**
- `amqp.Queue` - Declared queue information
- `error` - Declaration error if any

**Example:**
```go
queue, err := client.DeclareQueue("my_queue", true, false, false, nil)
if err != nil {
    log.Printf("Failed to declare queue: %v", err)
    return
}

log.Printf("Queue declared: %s", queue.Name)
```

### DeclareExchange

Declares an exchange on RabbitMQ.

```go
func (c *Client) DeclareExchange(
    name, kind string,
    durable, autoDelete, internal bool,
    args amqp.Table,
) error
```

**Parameters:**
- `name` - Exchange name
- `kind` - Exchange type (direct, fanout, topic, headers)
- `durable` - Whether the exchange survives broker restart
- `autoDelete` - Whether the exchange is deleted when last queue unsubscribes
- `internal` - Whether the exchange is internal
- `args` - Additional exchange arguments

**Returns:**
- `error` - Declaration error if any

**Example:**
```go
err := client.DeclareExchange("my_exchange", "direct", true, false, false, nil)
if err != nil {
    log.Printf("Failed to declare exchange: %v", err)
    return
}

log.Println("Exchange declared successfully")
```

### QueueBind

Binds a queue to an exchange with a routing key.

```go
func (c *Client) QueueBind(
    name, key, exchange string,
    noWait bool,
    args amqp.Table,
) error
```

**Parameters:**
- `name` - Queue name
- `key` - Routing key
- `exchange` - Exchange name
- `noWait` - Whether to wait for server confirmation
- `args` - Additional binding arguments

**Returns:**
- `error` - Binding error if any

**Example:**
```go
err := client.QueueBind("my_queue", "my_key", "my_exchange", false, nil)
if err != nil {
    log.Printf("Failed to bind queue: %v", err)
    return
}

log.Println("Queue bound successfully")
```

## Configuration

### Config

Client configuration structure.

```go
type Config struct {
    URLs                []string      // List of RabbitMQ connection URLs
    ReconnectInterval   time.Duration // Time between reconnection attempts
    MaxReconnectAttempt int           // Maximum number of reconnection attempts
    DebugLog            bool          // Enable/disable debug logging
    Logger              Logger        // Custom logger implementation
}
```

**Fields:**
- `URLs` - List of RabbitMQ connection URLs (default: `["amqp://localhost:5672"]`)
- `ReconnectInterval` - Time between reconnection attempts (default: `5s`)
- `MaxReconnectAttempt` - Maximum number of reconnection attempts (default: `10`)
- `DebugLog` - Enable/disable debug logging (default: `false`)
- `Logger` - Custom logger implementation (default: `DefaultLogger`)

## Error Handling

The client automatically handles connection errors and attempts reconnection based on the configuration. Common error scenarios:

- **Connection refused** - RabbitMQ server is not running
- **Authentication failed** - Invalid credentials
- **Channel error** - AMQP channel operation failed
- **Network timeout** - Connection timeout

## Best Practices

1. **Always close the client** using `defer client.Close()`
2. **Check connection status** before operations using `IsConnected()`
3. **Handle errors gracefully** and implement retry logic for critical operations
4. **Use context** for connection operations to support cancellation
5. **Monitor connection health** in production environments

## Examples

See [Basic Examples](../examples/basic.md) and [Advanced Examples](../examples/advanced.md) for complete working examples. 