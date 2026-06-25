package event

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"backend/internal/logger"
)

const defaultExchange = "cybertown.events"

// Publisher 向 RabbitMQ 发布事件，支持自动重连。
type Publisher struct {
	ch     *amqp.Channel
	conn   *amqp.Connection
	dsn    string
	appLog *logger.AppLogger
	mu     sync.Mutex
}

// NewPublisher 创建事件发布器，同时声明 exchange 和 queue。
func NewPublisher(ch *amqp.Channel, conn *amqp.Connection, dsn string, appLog *logger.AppLogger) (*Publisher, error) {
	p := &Publisher{ch: ch, conn: conn, dsn: dsn, appLog: appLog}
	if err := p.setupChannel(ch); err != nil {
		return nil, err
	}
	// Monitor channel close for auto-reconnect
	go p.watchChannel(ch)
	return p, nil
}

func (p *Publisher) setupChannel(ch *amqp.Channel) error {
	if err := ch.Confirm(false); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(
		defaultExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return err
	}

	tickQueues := []string{"town_tick_event", "town_tick_activity", "town_tick_interaction", "town_tick_story"}
	for _, q := range tickQueues {
		if _, err := ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(q, "town.*", defaultExchange, false, nil); err != nil {
			return err
		}
	}

	if _, err := ch.QueueDeclare("town_broadcast", true, false, false, false, nil); err != nil {
		return err
	}
	_ = ch.QueueUnbind("town_broadcast", "npc.*", defaultExchange, nil)
	_ = ch.QueueUnbind("town_broadcast", "town.*", defaultExchange, nil)

	if _, err := ch.QueueDeclare("npc_events", true, false, false, false, nil); err != nil {
		return err
	}
	for _, rk := range []string{EventTypeNPCMoveRequest, EventTypeNPCMoved} {
		if err := ch.QueueBind("npc_events", rk, defaultExchange, false, nil); err != nil {
			return err
		}
	}
	broadcastKeys := []string{
		EventTypeNPCMoved, EventTypeNPCReplied, EventTypeNPCIdleAction,
		EventTypeNPCActivityGenerated, EventTypeNPCMoodChanged,
		EventTypeNPCInteractionGenerated, EventTypeNPCRelationshipChanged,
		EventTypeStoryEventTriggered, EventTypeNPCGoalChanged,
		EventTypeTownNewsGenerated,
	}
	for _, rk := range broadcastKeys {
		if err := ch.QueueBind("town_broadcast", rk, defaultExchange, false, nil); err != nil {
			return err
		}
	}

	if _, err := ch.QueueDeclare("user_events", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind("user_events", "user.#", defaultExchange, false, nil); err != nil {
		return err
	}
	return nil
}

func (p *Publisher) watchChannel(ch *amqp.Channel) {
	closeCh := make(chan *amqp.Error, 1)
	ch.NotifyClose(closeCh)
	err := <-closeCh
	if err != nil {
		p.appLog.Warn("RabbitMQ publish channel closed, attempting reconnect", "reason", err.Reason)
	}
	p.reconnect()
}

func (p *Publisher) reconnect() {
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		conn, err := amqp.Dial(p.dsn)
		if err != nil {
			p.appLog.Warn("RabbitMQ reconnect dial failed, retrying...", "attempt", i+1)
			continue
		}
		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			p.appLog.Warn("RabbitMQ reconnect channel failed, retrying...", "attempt", i+1)
			continue
		}
		if err := p.setupChannel(ch); err != nil {
			ch.Close()
			conn.Close()
			p.appLog.Warn("RabbitMQ reconnect setup failed, retrying...", "attempt", i+1)
			continue
		}
		p.mu.Lock()
		oldConn := p.conn
		p.conn = conn
		p.ch = ch
		p.mu.Unlock()
		if oldConn != nil {
			oldConn.Close()
		}
		p.appLog.Info("RabbitMQ publisher reconnected successfully")
		go p.watchChannel(ch)
		return
	}
	p.appLog.Error(nil, "RabbitMQ publisher reconnect failed after 30 attempts")
}

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
