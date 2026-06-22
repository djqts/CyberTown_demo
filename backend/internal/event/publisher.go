package event

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/internal/logger"
)

const defaultExchange = "cybertown.events"

// Publisher 向 RabbitMQ 发布事件。
type Publisher struct {
	ch     *amqp.Channel
	appLog *logger.AppLogger
	mu     sync.Mutex
}

// NewPublisher 创建事件发布器，同时声明 exchange 和 queue。
func NewPublisher(ch *amqp.Channel, appLog *logger.AppLogger) (*Publisher, error) {
	if err := ch.ExchangeDeclare(
		defaultExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return nil, err
	}

	_, err := ch.QueueDeclare(
		"town_events", true, false, false, false, nil,
	)
	if err != nil {
		return nil, err
	}

	if _, err := ch.QueueDeclare(
		"town_broadcast", true, false, false, false, nil,
	); err != nil {
		return nil, err
	}

	_ = ch.QueueUnbind("town_events", "npc.*", defaultExchange, nil)
	_ = ch.QueueUnbind("town_broadcast", "npc.*", defaultExchange, nil)
	_ = ch.QueueUnbind("town_broadcast", "town.*", defaultExchange, nil)

	if err := ch.QueueBind("town_events", "town.*", defaultExchange, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("town_events", eventTypeRoutingKey(EventTypeNPCMoveRequest), defaultExchange, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("town_events", eventTypeRoutingKey(EventTypeNPCMoved), defaultExchange, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("town_broadcast", eventTypeRoutingKey(EventTypeNPCMoved), defaultExchange, false, nil); err != nil {
		return nil, err
	}
	if err := ch.QueueBind("town_broadcast", eventTypeRoutingKey(EventTypeNPCReplied), defaultExchange, false, nil); err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"user_events", true, false, false, false, nil,
	)
	if err != nil {
		return nil, err
	}
	if err := ch.QueueBind("user_events", "user.#", defaultExchange, false, nil); err != nil {
		return nil, err
	}

	return &Publisher{ch: ch, appLog: appLog}, nil
}

func eventTypeRoutingKey(eventType string) string {
	return eventType
}

// Publish 发布事件到 exchange，routing key 取 event_type。
func (p *Publisher) Publish(ctx context.Context, e *Event) error {
	body, err := Marshal(e)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.ch.PublishWithContext(ctx,
		defaultExchange,
		e.EventType,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		p.appLog.Error(err, "事件发布失败", "event_type", e.EventType, "event_id", e.EventID)
		return err
	}

	p.appLog.Info("事件已发布", "event_type", e.EventType, "event_id", e.EventID)
	return nil
}
