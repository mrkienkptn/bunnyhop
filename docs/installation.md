# Installation Guide

This guide will help you install and set up BunnyHop in your Go project.

## Prerequisites

Before installing BunnyHop, ensure you have:

- **Go 1.24 or later** - [Download Go](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Docker & Docker Compose** - For running RabbitMQ (optional but recommended)

## Installation Methods

### Method 1: Go Modules (Recommended)

Add BunnyHop to your Go module:

```bash
go get github.com/vanduc0209/bunnyhop
```

Then import it in your Go code:

```go
import "github.com/vanduc0209/bunnyhop"
```

### Method 2: Clone Repository

Clone the repository and use it locally:

```bash
git clone https://github.com/vanduc0209/bunnyhop.git
cd bunnyhop
go mod download
```

## Dependencies

BunnyHop has the following dependencies:

- **github.com/rabbitmq/amqp091-go** - AMQP 0.9.1 protocol implementation
- **github.com/stretchr/testify** - Testing framework (for development)

These will be automatically downloaded when you run `go mod download`.

## Setting Up RabbitMQ

### Option 1: Docker (Recommended for Development)

#### Single Node Setup

```bash
# Start a single RabbitMQ node
docker-compose -f docker/docker-compose.single.yml up -d

# Or use the provided script
chmod +x scripts/start-single.sh
./scripts/start-single.sh
```

#### Cluster Setup

```bash
# Start a 3-node cluster
docker-compose -f docker/docker-compose.cluster.yml up -d

# Or use the provided script
chmod +x scripts/start-cluster.sh
./scripts/start-cluster.sh
```

### Option 2: Local Installation

1. Download RabbitMQ from [rabbitmq.com](https://www.rabbitmq.com/download.html)
2. Install and start the service
3. Enable the management plugin: `rabbitmq-plugins enable rabbitmq_management`

### Option 3: Cloud Services

- **RabbitMQ Cloud** - [cloud.rabbitmq.com](https://www.cloudamqp.com/)
- **AWS MQ** - Managed RabbitMQ service
- **Azure Service Bus** - Alternative messaging service

## Verification

### Test Your Installation

Create a simple test file to verify everything works:

```go
package main

import (
    "log"
    "github.com/vanduc0209/bunnyhop"
)

func main() {
    config := bunnyhop.Config{
        URLs: []string{"amqp://guest:guest@localhost:5672/"},
        DebugLog: true,
    }
    
    client := bunnyhop.NewClient(config)
    log.Printf("BunnyHop client created successfully: %v", client)
}
```

Run the test:

```bash
go run main.go
```

### Check RabbitMQ

If using Docker, verify RabbitMQ is running:

```bash
# Check container status
docker ps

# Access management UI
# Single node: http://localhost:15672 (guest/guest)
# Cluster: http://localhost:15670 (via HAProxy)
```

## Configuration

### Environment Variables

Set these environment variables for production use:

```bash
export RABBITMQ_URL="amqp://user:pass@host:5672/"
export RABBITMQ_DEBUG="true"
export RABBITMQ_RECONNECT_INTERVAL="5s"
export RABBITMQ_MAX_RECONNECT_ATTEMPTS="10"
```

### Configuration Files

Create a configuration file for your application:

```yaml
# config.yaml
rabbitmq:
  urls:
    - "amqp://user:pass@node1:5672/"
    - "amqp://user:pass@node2:5672/"
  reconnect_interval: "5s"
  max_reconnect_attempts: 10
  debug_log: false
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure RabbitMQ is running
   - Check port accessibility
   - Verify credentials

2. **Import Errors**
   - Run `go mod tidy`
   - Check Go version compatibility
   - Verify module path

3. **Docker Issues**
   - Ensure Docker is running
   - Check port conflicts
   - Verify Docker Compose version

### Getting Help

- Check the [Troubleshooting Guide](troubleshooting.md)
- Review [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- Open an issue on [GitHub](https://github.com/vanduc0209/bunnyhop/issues)

## Next Steps

Now that you have BunnyHop installed:

1. Read the [Quick Start Guide](quickstart.md) to begin using the library
2. Explore [Basic Examples](examples/basic.md) for common use cases
3. Review [Configuration Options](configuration.md) for advanced setup
4. Check [Production Deployment](production.md) for production best practices 