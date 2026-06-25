package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/repo"
	"backend/internal/story"
)

type StoryWorker struct {
	consumer  consumer
	publisher eventPublisher
	eventRepo eventLogCreator
	storySvc  *story.Service
	npcRepo   npcStatusUpdater
	appLog    *logger.AppLogger
}

func NewStoryWorker(
	consumer consumer,
	publisher eventPublisher,
	eventRepo eventLogCreator,
	storyRepo *repo.StoryRepo,
	npcRepo npcStatusUpdater,
	appLog *logger.AppLogger,
) *StoryWorker {
	return &StoryWorker{
		consumer:  consumer,
		publisher: publisher,
		eventRepo: eventRepo,
		storySvc:  story.NewService(storyRepo, npcRepo, appLog),
		npcRepo:   npcRepo,
		appLog:    appLog,
	}
}

func (w *StoryWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "town_tick_story", w.handleEvent)
}

func (w *StoryWorker) handleEvent(ctx context.Context, e *event.Event) error {
	if e.EventType != event.EventTypeTownTick {
		return nil
	}

	var tick struct {
		Day    int `json:"day"`
		Minute int `json:"minute"`
	}
	if err := json.Unmarshal(e.Payload, &tick); err != nil {
		return fmt.Errorf("StoryWorker parse tick: %w", err)
	}

	effects := w.storySvc.CheckAndTrigger(e.TownID, tick.Minute, tick.Day)

	if len(effects) > 0 {
		// Publish town news for the story event
		w.publishTownNews(ctx, e, effects[0].StoryTitle)
	}

	for _, eff := range effects {
		w.publishStoryTriggered(ctx, e, &eff)

		if eff.NewMood != eff.OldMood {
			_ = w.npcRepo.UpdateMood(eff.NPCID, eff.NewMood)
			w.publishMoodChanged(ctx, e, &eff)
		}

		if eff.NewGoal != eff.OldGoal {
			_ = w.npcRepo.UpdateGoal(eff.NPCID, eff.NewGoal)
			w.publishGoalChanged(ctx, e, &eff)
		}
	}

	return nil
}

func (w *StoryWorker) publishStoryTriggered(ctx context.Context, tickEvent *event.Event, eff *story.TriggeredEffect) {
	payload, _ := json.Marshal(eff)
	se := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeStoryEventTriggered,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeSystem,
		ActorID:   "story_system",
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}
	if err := w.publisher.Publish(ctx, se); err != nil {
		w.appLog.Error(err, "publish story.event.triggered failed")
	}
	_ = writeEventLog(w.eventRepo, se)
}

func (w *StoryWorker) publishMoodChanged(ctx context.Context, tickEvent *event.Event, eff *story.TriggeredEffect) {
	payload, _ := json.Marshal(map[string]any{
		"npc_id":   eff.NPCID,
		"npc_name": eff.NPCName,
		"old_mood": eff.OldMood,
		"new_mood": eff.NewMood,
		"reason":   eff.StoryTitle,
	})
	me := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCMoodChanged,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", eff.NPCID),
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}
	if err := w.publisher.Publish(ctx, me); err != nil {
		w.appLog.Error(err, "publish npc.mood.changed failed")
	}
	_ = writeEventLog(w.eventRepo, me)
}

func (w *StoryWorker) publishGoalChanged(ctx context.Context, tickEvent *event.Event, eff *story.TriggeredEffect) {
	payload, _ := json.Marshal(map[string]any{
		"npc_id":   eff.NPCID,
		"npc_name": eff.NPCName,
		"old_goal": eff.OldGoal,
		"new_goal": eff.NewGoal,
		"reason":   eff.StoryTitle,
	})
	ge := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCGoalChanged,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", eff.NPCID),
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}
	if err := w.publisher.Publish(ctx, ge); err != nil {
		w.appLog.Error(err, "publish npc.goal.changed failed")
	}
	_ = writeEventLog(w.eventRepo, ge)
}

func (w *StoryWorker) publishTownNews(ctx context.Context, tickEvent *event.Event, storyTitle string) {
	payload, _ := json.Marshal(map[string]any{
		"story_title": storyTitle,
		"message":     "小镇新闻：" + storyTitle,
	})
	ne := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeTownNewsGenerated,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeSystem,
		ActorID:   "news_system",
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}
	if err := w.publisher.Publish(ctx, ne); err != nil {
		w.appLog.Error(err, "publish town.news.generated failed")
	}
	_ = writeEventLog(w.eventRepo, ne)
}
