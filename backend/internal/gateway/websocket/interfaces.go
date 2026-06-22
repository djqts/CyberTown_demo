package websocket

import (
	"context"

	"backend/internal/event"
)

type eventPublisher interface {
	Publish(ctx context.Context, e *event.Event) error
}
