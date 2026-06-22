package infra

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	config "backend/internal/config"
	"backend/internal/logger"
)

// RabbitMQClient 包装 *amqp.Connection。
type RabbitMQClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitMQClient 创建并验证 RabbitMQ 连接。
func NewRabbitMQClient(appLog *logger.AppLogger) (*RabbitMQClient, error) {
	cfg := config.AppConfig.RabbitMQ
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	appLog.Info("正在连接 RabbitMQ", "host", cfg.Host, "port", cfg.Port, "user", cfg.User)

	conn, err := amqp.Dial(dsn)
	if err != nil {
		appLog.Error(err, "RabbitMQ 连接失败", "host", cfg.Host)
		return nil, fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		appLog.Error(err, "RabbitMQ 创建 Channel 失败")
		return nil, fmt.Errorf("amqp channel: %w", err)
	}

	appLog.Info("RabbitMQ 连接成功", "host", cfg.Host)
	return &RabbitMQClient{Conn: conn, Channel: ch}, nil
}

// Close 关闭 Channel 和 Connection。
func (r *RabbitMQClient) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}
