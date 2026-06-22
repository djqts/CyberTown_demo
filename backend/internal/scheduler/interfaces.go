package scheduler

import (
	"context"

	"backend/internal/event"
	"backend/internal/model"
)

type publisher interface {
	Publish(ctx context.Context, e *event.Event) error
}

type townTimeAdvancer interface {
	AdvanceTime(townID uint) (*model.Town, bool, error)
}
