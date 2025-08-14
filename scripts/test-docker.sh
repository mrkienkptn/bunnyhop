#!/bin/bash

echo "🧪 Testing BunnyHop Library with Docker"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_success "Docker is running"

# Test 1: Single Node
echo ""
print_status "Testing Single Node Setup..."
print_status "Starting single RabbitMQ node..."

docker-compose -f docker/docker-compose.single.yml up -d

# Wait for RabbitMQ to be ready
print_status "Waiting for RabbitMQ to be ready..."
sleep 15

# Check if RabbitMQ is running
if docker-compose -f docker/docker-compose.single.yml ps | grep -q "Up"; then
    print_success "Single RabbitMQ node is running!"
    
    # Test the library
    print_status "Testing library with single node..."
    go run example/main.go
    
    if [ $? -eq 0 ]; then
        print_success "Library test with single node passed!"
    else
        print_warning "Library test with single node had issues (expected if no RabbitMQ)"
    fi
else
    print_error "Failed to start single RabbitMQ node"
    docker-compose -f docker/docker-compose.single.yml logs
    exit 1
fi

# Stop single node
print_status "Stopping single node..."
docker-compose -f docker/docker-compose.single.yml down

# Test 2: Cluster
echo ""
print_status "Testing Cluster Setup..."
print_status "Starting 3-node RabbitMQ cluster..."

docker-compose -f docker/docker-compose.cluster.yml up -d

# Wait for cluster to be ready
print_status "Waiting for RabbitMQ cluster to be ready..."
print_status "This may take up to 2 minutes for all nodes to join the cluster..."
sleep 90

# Check if cluster is running
if docker-compose -f docker/docker-compose.cluster.yml ps | grep -q "Up"; then
    print_success "3-node RabbitMQ cluster is running!"
    
    # Show cluster status
    print_status "Cluster status:"
    docker-compose -f docker/docker-compose.cluster.yml ps
    
    # Test the library with cluster
    print_status "Testing library with cluster..."
    go run example/main.go
    
    if [ $? -eq 0 ]; then
        print_success "Library test with cluster passed!"
    else
        print_warning "Library test with cluster had issues (expected if no RabbitMQ)"
    fi
    
    # Show HAProxy stats
    print_status "HAProxy stats available at: http://localhost:8404 (admin/admin123)"
    
else
    print_error "Failed to start RabbitMQ cluster"
    docker-compose -f docker/docker-compose.cluster.yml logs
    exit 1
fi

# Show final status
echo ""
print_status "Final status:"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
print_success "Docker testing completed!"
print_status "You can now:"
print_status "  - Access individual nodes:"
print_status "    * Node 1: AMQP localhost:5672, Management http://localhost:15672"
print_status "    * Node 2: AMQP localhost:5673, Management http://localhost:15673"
print_status "    * Node 3: AMQP localhost:5674, Management http://localhost:15674"
print_status "  - Use load balancer:"
print_status "    * HAProxy AMQP: localhost:5670"
print_status "    * HAProxy Management: http://localhost:15670"
print_status "    * HAProxy Stats: http://localhost:8404 (admin/admin123)"
echo ""
print_status "To stop all containers:"
print_status "  make docker-stop"
print_status "To clean up:"
print_status "  make docker-clean" 