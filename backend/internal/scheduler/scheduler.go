package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/event"
	"backend/internal/logger"
)

// Scheduler 定时推进小镇时间并发布事件。
type Scheduler struct {
	interval  time.Duration
	townID    uint
	publisher publisher
	townSvc   townTimeAdvancer
	appLog    *logger.AppLogger
	stopCh    chan struct{}
}

// New 创建调度器。interval 为推进时间间隔。
func New(
	interval time.Duration,
	townID uint,
	publisher *event.Publisher,
	townSvc townTimeAdvancer,
	appLog *logger.AppLogger,
) *Scheduler {
	return &Scheduler{
		interval:  interval,
		townID:    townID,
		publisher: publisher,
		townSvc:   townSvc,
		appLog:    appLog,
		stopCh:    make(chan struct{}),
	}
}

// Start 启动定时器，每 interval 推进一次小镇时间。
func (s *Scheduler) Start(ctx context.Context) {
	s.appLog.Info("调度器启动", "interval", s.interval.String(), "town_id", s.townID)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.appLog.Info("调度器已停止")
			return
		case <-s.stopCh:
			s.appLog.Info("调度器已停止")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

// Stop 停止调度器。
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) tick(ctx context.Context) {
	town, dayRolled, err := s.townSvc.AdvanceTime(s.townID)
	if err != nil {
		s.appLog.Error(err, "推进时间失败", "town_id", s.townID)
		return
	}

	s.appLog.Info("小镇时间推进",
		"day", town.CurrentDay,
		"minute", town.CurrentMinute,
		"town_id", town.ID,
	)

	payload, _ := json.Marshal(map[string]any{
		"day":        town.CurrentDay,
		"minute":     town.CurrentMinute,
		"day_rolled": dayRolled,
	})

	e := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeTownTick,
		TownID:    town.ID,
		ActorType: event.ActorTypeSystem,
		ActorID:   "scheduler",
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	if err := s.publisher.Publish(ctx, e); err != nil {
		s.appLog.Error(err, "发布 town.tick 失败")
	}
}

func newEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
