package story

import (
	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/repo"
)

type NPCByNameFinder interface {
	FindByTownID(townID uint) ([]model.NPC, error)
}

type Service struct {
	storyRepo *repo.StoryRepo
	npcRepo   NPCByNameFinder
	appLog    *logger.AppLogger
}

func NewService(storyRepo *repo.StoryRepo, npcRepo NPCByNameFinder, appLog *logger.AppLogger) *Service {
	return &Service{
		storyRepo: storyRepo,
		npcRepo:   npcRepo,
		appLog:    appLog,
	}
}

func (s *Service) CheckAndTrigger(townID uint, minuteOfDay, day int) []TriggeredEffect {
	events, err := s.storyRepo.FindByTownID(townID)
	if err != nil {
		s.appLog.Error(err, "find story events failed")
		return nil
	}

	npcs, err := s.npcRepo.FindByTownID(townID)
	if err != nil {
		s.appLog.Error(err, "find npcs for story effects failed")
		return nil
	}

	npcByName := make(map[string]*model.NPC, len(npcs))
	for i := range npcs {
		npcByName[npcs[i].Name] = &npcs[i]
	}

	var triggered []TriggeredEffect
	for _, evt := range events {
		if evt.Status != "ready" {
			continue
		}

		if !ShouldTrigger(evt.TriggerCondition, minuteOfDay, evt.LastTriggeredAt) {
			continue
		}

		s.appLog.Info("story event triggered", "title", evt.Title)

		effects, err := ParseEffects(evt.Effects)
		if err != nil {
			s.appLog.Error(err, "parse story effects failed", "title", evt.Title)
			continue
		}

		for _, eff := range effects.NPCEffects {
			npc, ok := npcByName[eff.NPCName]
			if !ok {
				s.appLog.Warn("NPC not found for story effect", "npc_name", eff.NPCName)
				continue
			}

			te := TriggeredEffect{
				StoryTitle: evt.Title,
				NPCID:      npc.ID,
				NPCName:    npc.Name,
				OldMood:    npc.Mood,
				NewMood:    npc.Mood,
				OldGoal:    npc.CurrentGoal,
				NewGoal:    npc.CurrentGoal,
			}
			if eff.Mood != "" {
				te.NewMood = eff.Mood
			}
			if eff.Goal != "" {
				te.NewGoal = eff.Goal
			}
			triggered = append(triggered, te)
		}
		_ = s.storyRepo.MarkTriggered(evt.ID)
	}
	return triggered
}

type TriggeredEffect struct {
	StoryTitle string `json:"story_title"`
	NPCID      uint   `json:"npc_id"`
	NPCName    string `json:"npc_name"`
	OldMood    string `json:"old_mood"`
	NewMood    string `json:"new_mood"`
	OldGoal    string `json:"old_goal"`
	NewGoal    string `json:"new_goal"`
}
