package bunnyhop

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config cấu hình cho Client
type Config struct {
	URLs                []string      // Danh sách URLs của RabbitMQ
	ReconnectInterval   time.Duration // Thời gian chờ giữa các lần reconnect
	MaxReconnectAttempt int           // Số lần thử reconnect tối đa
	DebugLog            bool          // Bật/tắt debug log
	Logger              Logger        // Custom logger interface
}

// Client quản lý kết nối đến RabbitMQ
type Client struct {
	config            Config
	connection        *amqp.Connection
	channel           *amqp.Channel
	mutex             sync.RWMutex
	connected         bool
	reconnectAttempts int
	ctx               context.Context
	cancel            context.CancelFunc
	reconnectTicker   *time.Ticker
	connectionErrors  chan *amqp.Error
	channelErrors     chan *amqp.Error
}

// NewClient tạo client mới
func NewClient(config Config) *Client {
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.MaxReconnectAttempt == 0 {
		config.MaxReconnectAttempt = 10
	}
	if config.Logger == nil {
		config.Logger = NewDefaultLogger(config.DebugLog)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		config:           config,
		ctx:              ctx,
		cancel:           cancel,
		connectionErrors: make(chan *amqp.Error, 1),
		channelErrors:    make(chan *amqp.Error, 1),
	}
}

// Connect thiết lập kết nối đến RabbitMQ
func (c *Client) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return nil
	}

	c.logger().Debug("Connecting to RabbitMQ...")

	// Thử kết nối đến từng URL
	var lastErr error
	for _, url := range c.config.URLs {
		if err := c.connectToURL(url); err != nil {
			lastErr = err
			c.logger().Warn("Failed to connect to %s: %v", url, err)
			continue
		}
		c.logger().Info("Successfully connected to %s", url)
		return nil
	}

	return fmt.Errorf("failed to connect to any RabbitMQ server: %v", lastErr)
}

// connectToURL kết nối đến một URL cụ thể
func (c *Client) connectToURL(url string) error {
	// Tạo connection
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	// Tạo channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %v", err)
	}

	// Thiết lập QoS
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	// Lưu connection và channel
	c.connection = conn
	c.channel = ch
	c.connected = true
	c.reconnectAttempts = 0

	// Thiết lập error handlers
	c.setupErrorHandlers()

	// Bắt đầu reconnect goroutine
	go c.reconnectWorker()

	return nil
}

// setupErrorHandlers thiết lập xử lý lỗi
func (c *Client) setupErrorHandlers() {
	if c.connection != nil {
		c.connectionErrors = c.connection.NotifyClose(make(chan *amqp.Error, 1))
	}
	if c.channel != nil {
		c.channelErrors = c.channel.NotifyClose(make(chan *amqp.Error, 1))
	}
}

// reconnectWorker xử lý reconnect tự động
func (c *Client) reconnectWorker() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.connectionErrors:
			if err != nil {
				c.logger().Error("Connection error: %v", err)
				c.handleDisconnection()
			}
		case err := <-c.channelErrors:
			if err != nil {
				c.logger().Error("Channel error: %v", err)
				c.handleDisconnection()
			}
		}
	}
}

// handleDisconnection xử lý khi mất kết nối
func (c *Client) handleDisconnection() {
	c.mutex.Lock()
	c.connected = false
	c.mutex.Unlock()

	c.logger().Warn("Connection lost, attempting to reconnect...")

	// Thử reconnect
	go c.reconnect()
}

// reconnect thực hiện reconnect
func (c *Client) reconnect() {
	c.mutex.Lock()

	if c.connected {
		c.mutex.Unlock()
		return
	}

	c.reconnectAttempts++
	if c.reconnectAttempts > c.config.MaxReconnectAttempt {
		c.logger().Error("Max reconnection attempts reached")
		c.mutex.Unlock()
		return
	}

	c.logger().Info("Reconnection attempt %d/%d", c.reconnectAttempts, c.config.MaxReconnectAttempt)

	// Đóng connection cũ nếu có
	if c.connection != nil {
		c.connection.Close()
		c.connection = nil
	}
	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}

	c.mutex.Unlock()
	// Thử kết nối lại
	if err := c.Connect(c.ctx); err != nil {
		c.logger().Error("Reconnection failed: %v", err)
		// Thử lại sau một khoảng thời gian
		time.AfterFunc(c.config.ReconnectInterval, c.reconnect)
	}
}

// IsConnected kiểm tra trạng thái kết nối
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected && c.connection != nil && !c.connection.IsClosed()
}

// GetChannel lấy channel hiện tại
func (c *Client) GetChannel() (*amqp.Channel, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected || c.channel == nil || c.channel.IsClosed() {
		return nil, fmt.Errorf("client is not connected")
	}

	return c.channel, nil
}

// GetConnection lấy connection hiện tại
func (c *Client) GetConnection() (*amqp.Connection, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.connected || c.connection != nil || c.connection.IsClosed() {
		return nil, fmt.Errorf("client is not connected")
	}

	return c.connection, nil
}

// Close đóng kết nối
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cancel()

	if c.reconnectTicker != nil {
		c.reconnectTicker.Stop()
	}

	var errs []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %v", err))
		}
		c.channel = nil
	}

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %v", err))
		}
		c.connection = nil
	}

	c.connected = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// logger helper để lấy logger
func (c *Client) logger() Logger {
	return c.config.Logger
}

// PublishMessage gửi message
func (c *Client) PublishMessage(
	exchange, routingKey string,
	mandatory, immediate bool,
	msg amqp.Publishing,
) error {
	ch, err := c.GetChannel()
	if err != nil {
		return err
	}

	return ch.Publish(exchange, routingKey, mandatory, immediate, msg)
}

// DeclareQueue khai báo queue
func (c *Client) DeclareQueue(
	name string,
	durable, autoDelete, exclusive bool,
	args amqp.Table,
) (amqp.Queue, error) {
	ch, err := c.GetChannel()
	if err != nil {
		return amqp.Queue{}, err
	}

	return ch.QueueDeclare(name, durable, autoDelete, exclusive, false, args)
}

// DeclareExchange khai báo exchange
func (c *Client) DeclareExchange(
	name, kind string,
	durable, autoDelete, internal bool,
	args amqp.Table,
) error {
	ch, err := c.GetChannel()
	if err != nil {
		return err
	}

	return ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, false, args)
}

// QueueBind bind queue với exchange
func (c *Client) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	ch, err := c.GetChannel()
	if err != nil {
		return err
	}

	return ch.QueueBind(name, key, exchange, noWait, args)
}
