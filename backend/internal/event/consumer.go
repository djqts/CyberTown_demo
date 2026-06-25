package event

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/internal/logger"
)

// Handler 是事件处理回调函数类型。
type Handler func(ctx context.Context, e *Event) error

// Consumer 从 RabbitMQ 消费事件。
type Consumer struct {
	ch     *amqp.Channel
	appLog *logger.AppLogger
}

// NewConsumer 创建事件消费者。
func NewConsumer(ch *amqp.Channel, appLog *logger.AppLogger) *Consumer {
	return &Consumer{ch: ch, appLog: appLog}
}

// Consume 开始从指定队列消费事件，每条消息调用 handler 处理。
func (c *Consumer) Consume(ctx context.Context, queue string, handler Handler) error {
	msgs, err := c.ch.ConsumeWithContext(ctx,
		queue, "", false, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	c.appLog.Info("开始消费事件", "queue", queue)

	for {
		select {
		case <-ctx.Done():
			c.appLog.Info("事件消费已停止", "queue", queue)
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			e, err := Unmarshal(msg.Body)
			if err != nil {
				c.appLog.Error(err, "事件解码失败", "body", string(msg.Body))
				msg.Nack(false, false)
				continue
			}
			if queue == "town_broadcast" {
				c.appLog.Info("事件已接收(broadcast)", "event_type", e.EventType, "event_id", e.EventID)
			}
			if err := handler(ctx, e); err != nil {
				c.appLog.Error(err, "事件处理失败", "event_type", e.EventType, "event_id", e.EventID)
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		}
	}
}
