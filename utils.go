package bunnyhop

import (
	"time"
)

func getDefaultConfig(config *PoolConfig) {
	if config == nil {
		config = &PoolConfig{}
	}

	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 30 * time.Second
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if len(config.URLs) == 0 {
		config.URLs = []string{"amqp://localhost:5672"}
	}
	if config.Logger == nil {
		config.Logger = NewDefaultLogger(config.DebugLog)
	}
}
