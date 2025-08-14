.PHONY: help test test-race build clean docker-single docker-cluster docker-stop docker-clean

# Default target
help:
	@echo "Available commands:"
	@echo "  test          - Run unit tests"
	@echo "  test-race     - Run tests with race detection"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-single - Start single RabbitMQ node"
	@echo "  docker-cluster- Start 3-node RabbitMQ cluster"
	@echo "  docker-stop   - Stop all Docker containers"
	@echo "  docker-clean  - Clean Docker containers and volumes"

# Run unit tests
test:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race -v ./...

# Build the application
build:
	go build -o bin/bunnyhop .

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Start single RabbitMQ node
docker-single:
	docker-compose -f docker-compose.single.yml up -d
	@echo "Single RabbitMQ node started at:"
	@echo "  AMQP: localhost:5672"
	@echo "  Management UI: http://localhost:15672 (guest/guest)"

# Start 3-node RabbitMQ cluster
docker-cluster:
	docker-compose -f docker-compose.cluster.yml up -d
	@echo "3-node RabbitMQ cluster started at:"
	@echo "  Node 1 - AMQP: localhost:5672, Management: http://localhost:15672"
	@echo "  Node 2 - AMQP: localhost:5673, Management: http://localhost:15673"
	@echo "  Node 3 - AMQP: localhost:5674, Management: http://localhost:15674"
	@echo "  HAProxy - AMQP: localhost:5670, Management: http://localhost:15670"
	@echo "  HAProxy Stats: http://localhost:8404 (admin/admin123)"

# Stop all Docker containers
docker-stop:
	docker-compose -f docker-compose.single.yml down
	docker-compose -f docker-compose.cluster.yml down

# Clean Docker containers and volumes
docker-clean:
	docker-compose -f docker-compose.single.yml down -v
	docker-compose -f docker-compose.cluster.yml down -v
	docker system prune -f

# Show running containers
docker-ps:
	docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Show logs for single node
docker-logs-single:
	docker-compose -f docker-compose.single.yml logs -f

# Show logs for cluster
docker-logs-cluster:
	docker-compose -f docker-compose.cluster.yml logs -f

# Test the library with single node
test-single: docker-single
	@echo "Waiting for RabbitMQ to be ready..."
	@sleep 10
	go run example/main.go

# Test the library with cluster
test-cluster: docker-cluster
	@echo "Waiting for RabbitMQ cluster to be ready..."
	@sleep 90
	go run example/main.go

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Run benchmarks
bench:
	go test -bench=. ./...

# Generate test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html" 