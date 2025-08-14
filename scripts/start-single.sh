#!/bin/bash

echo "🚀 Starting single RabbitMQ node..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.single.yml down

# Start single node
echo "📦 Starting RabbitMQ single node..."
docker-compose -f docker-compose.single.yml up -d

# Wait for RabbitMQ to be ready
echo "⏳ Waiting for RabbitMQ to be ready..."
sleep 10

# Check if RabbitMQ is running
if docker-compose -f docker-compose.single.yml ps | grep -q "Up"; then
    echo "✅ RabbitMQ single node is running!"
    echo ""
    echo "📋 Connection details:"
    echo "  AMQP: localhost:5672"
    echo "  Management UI: http://localhost:15672"
    echo "  Username: guest"
    echo "  Password: guest"
    echo ""
    echo "🔍 Check status: docker-compose -f docker-compose.single.yml ps"
    echo "📝 View logs: docker-compose -f docker-compose.single.yml logs -f"
    echo "🛑 Stop: docker-compose -f docker-compose.single.yml down"
else
    echo "❌ Failed to start RabbitMQ. Check logs:"
    docker-compose -f docker-compose.single.yml logs
    exit 1
fi 