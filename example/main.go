package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vanduc0209/bunnyhop"
)

func main() {
	// Tạo cấu hình cho pool
	config := bunnyhop.PoolConfig{
		URLs: []string{
			"amqp://guest:guest@localhost:5672/",
			"amqp://guest:guest@localhost:5673/",
			"amqp://guest:guest@localhost:5674/",
		},
		ReconnectInterval:   5 * time.Second,
		MaxReconnectAttempt: 10,
		HealthCheckInterval: 30 * time.Second,
		LoadBalanceStrategy: bunnyhop.RoundRobin,
		DebugLog:            true,
	}

	// Tạo pool mới
	pool := bunnyhop.NewPool(config)

	// Bắt đầu pool
	if err := pool.Start(); err != nil {
		log.Fatalf("Failed to start pool: %v", err)
	}

	// Đợi một chút để các connections được thiết lập
	time.Sleep(2 * time.Second)

	// Lấy client từ pool
	client, err := pool.GetClient()
	if err != nil {
		log.Printf("Failed to get client: %v", err)
	} else {
		log.Printf("Successfully got client")

		// Sử dụng client để khai báo queue
		queue, err := client.DeclareQueue("test_queue", true, false, false, nil)
		if err != nil {
			log.Printf("Failed to declare queue: %v", err)
		} else {
			log.Printf("Queue declared: %s", queue.Name)
		}
	}

	// Lấy thống kê của pool
	stats := pool.GetStats()
	fmt.Printf("Pool Stats: %+v\n", stats)

	// Đợi một chút để xem health check hoạt động
	time.Sleep(10 * time.Second)

	// Lấy thống kê mới
	stats = pool.GetStats()
	fmt.Printf("Updated Pool Stats: %+v\n", stats)

	// Đóng pool
	if err := pool.Close(); err != nil {
		log.Printf("Error closing pool: %v", err)
	}

	log.Println("Example completed")
}
