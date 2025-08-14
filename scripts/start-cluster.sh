#!/bin/bash

echo "🚀 Starting 3-node RabbitMQ cluster..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.cluster.yml down

# Start cluster
echo "📦 Starting RabbitMQ cluster..."
docker-compose -f docker-compose.cluster.yml up -d

# Wait for cluster to be ready
echo "⏳ Waiting for RabbitMQ cluster to be ready..."
echo "   This may take up to 2 minutes for all nodes to join the cluster..."
sleep 90

# Check if cluster is running
if docker-compose -f docker-compose.cluster.yml ps | grep -q "Up"; then
    echo "✅ RabbitMQ cluster is running!"
    echo ""
    echo "📋 Connection details:"
    echo "  Node 1 - AMQP: localhost:5672, Management: http://localhost:15672"
    echo "  Node 2 - AMQP: localhost:5673, Management: http://localhost:15673"
    echo "  Node 3 - AMQP: localhost:5674, Management: http://localhost:15674"
    echo ""
    echo "⚖️  Load Balancer (HAProxy):"
    echo "  AMQP: localhost:5670"
    echo "  Management UI: http://localhost:15670"
    echo "  Stats: http://localhost:8404 (admin/admin123)"
    echo ""
    echo "🔍 Check status: docker-compose -f docker-compose.cluster.yml ps"
    echo "📝 View logs: docker-compose -f docker-compose.cluster.yml logs -f"
    echo "🛑 Stop: docker-compose -f docker-compose.cluster.yml down"
    echo ""
    echo "🧪 Test the cluster: go run example/main.go"
else
    echo "❌ Failed to start RabbitMQ cluster. Check logs:"
    docker-compose -f docker-compose.cluster.yml logs
    exit 1
fi 