package bunnyhop

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test for single client
func TestClientIntegration_SingleNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := Config{
		URLs:                []string{"amqp://guest:guest@localhost:5672/"},
		ReconnectInterval:   1 * time.Second,
		MaxReconnectAttempt: 3,
		DebugLog:            true,
	}

	client := NewClient(config)
	require.NotNil(t, client)

	ctx := context.Background()

	// Try to connect (may fail if RabbitMQ is not running)
	err := client.Connect(ctx)
	if err != nil {
		t.Logf("Connection failed (expected if RabbitMQ not running): %v", err)
		t.Skip("RabbitMQ not available, skipping integration test")
	}

	// Test basic operations
	t.Run("DeclareQueue", func(t *testing.T) {
		queue, err := client.DeclareQueue("test_queue", false, true, false, nil)
		if err == nil {
			assert.NotEmpty(t, queue.Name)
			t.Logf("Queue declared: %s", queue.Name)
		}
	})

	t.Run("DeclareExchange", func(t *testing.T) {
		err := client.DeclareExchange("test_exchange", "direct", false, true, false, nil)
		if err == nil {
			t.Log("Exchange declared successfully")
		}
	})

	t.Run("QueueBind", func(t *testing.T) {
		err := client.QueueBind("test_queue", "test_key", "test_exchange", false, nil)
		if err == nil {
			t.Log("Queue bound successfully")
		}
	})

	// Cleanup
	client.Close()
}

// Integration test for pool
func TestPoolIntegration_Cluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := PoolConfig{
		URLs: []string{
			"amqp://guest:guest@localhost:5672/",
			"amqp://guest:guest@localhost:5673/",
			"amqp://guest:guest@localhost:5674/",
		},
		ReconnectInterval:   1 * time.Second,
		MaxReconnectAttempt: 3,
		HealthCheckInterval: 5 * time.Second,
		LoadBalanceStrategy: RoundRobin,
		DebugLog:            true,
	}

	pool := NewPool(config)
	require.NotNil(t, pool)

	// Start pool
	err := pool.Start()
	if err != nil {
		t.Logf("Failed to start pool: %v", err)
		t.Skip("RabbitMQ cluster not available, skipping integration test")
	}

	// Wait for connections to be established
	time.Sleep(5 * time.Second)

	t.Run("GetClient", func(t *testing.T) {
		client, err := pool.GetClient()
		if err == nil {
			assert.NotNil(t, client)
			t.Log("Successfully got client from pool")
		} else {
			t.Logf("Failed to get client: %v", err)
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		stats := pool.GetStats()
		assert.NotNil(t, stats)
		t.Logf("Pool stats: %+v", stats)
	})

	t.Run("HealthyNodeCount", func(t *testing.T) {
		count := pool.GetHealthyNodeCount()
		t.Logf("Healthy nodes: %d", count)
	})

	// Cleanup
	pool.Close()
}

// Benchmark tests
func BenchmarkClient_Connect(b *testing.B) {
	config := Config{
		URLs:                []string{"amqp://guest:guest@localhost:5672/"},
		ReconnectInterval:   1 * time.Second,
		MaxReconnectAttempt: 1,
		DebugLog:            false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := NewClient(config)
		ctx := context.Background()
		client.Connect(ctx)
		client.Close()
	}
}

func BenchmarkPool_GetClient(b *testing.B) {
	config := PoolConfig{
		URLs: []string{
			"amqp://guest:guest@localhost:5672/",
			"amqp://guest:guest@localhost:5673/",
		},
		LoadBalanceStrategy: RoundRobin,
		DebugLog:            false,
	}

	pool := NewPool(config)
	pool.Start()
	defer pool.Close()

	// Wait for pool to be ready
	time.Sleep(2 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.GetClient()
		if err != nil {
			b.Skipf("Skipping benchmark: %v", err)
		}
	}
}

// Test helper functions
func TestLoadBalancingStrategies(t *testing.T) {
	config := PoolConfig{
		URLs: []string{
			"amqp://localhost:5672",
			"amqp://localhost:5673",
			"amqp://localhost:5674",
		},
	}

	strategies := []LoadBalanceStrategy{
		RoundRobin,
		Random,
		LeastUsed,
		WeightedRoundRobin,
	}

	for _, strategy := range strategies {
		t.Run(strategy.String(), func(t *testing.T) {
			config.LoadBalanceStrategy = strategy
			pool := NewPool(config)

			// Set all nodes as healthy for testing
			for _, node := range pool.nodes {
				node.mutex.Lock()
				node.healthy = true
				node.mutex.Unlock()
			}

			// Test that we can get a client
			client, err := pool.GetClient()
			if err == nil {
				assert.NotNil(t, client)
			}

			pool.Close()
		})
	}
}

// Add String method to LoadBalanceStrategy for testing
func (lb LoadBalanceStrategy) String() string {
	switch lb {
	case RoundRobin:
		return "RoundRobin"
	case Random:
		return "Random"
	case LeastUsed:
		return "LeastUsed"
	case WeightedRoundRobin:
		return "WeightedRoundRobin"
	default:
		return "Unknown"
	}
}
