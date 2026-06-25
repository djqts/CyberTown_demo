package interaction

import (
	"context"

	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/repo"
)

type npcByLocationFinder interface {
	FindByTownID(townID uint) ([]model.NPC, error)
}

type Service struct {
	matcher      *InteractionMatcher
	relRepo      *repo.RelationshipRepo
	npcRepo      npcByLocationFinder
	socialSpread *SocialPropagation
	appLog       *logger.AppLogger
}

func NewService(relRepo *repo.RelationshipRepo, npcRepo npcByLocationFinder, appLog *logger.AppLogger) *Service {
	return &Service{
		matcher:      NewInteractionMatcher(),
		relRepo:      relRepo,
		npcRepo:      npcRepo,
		socialSpread: NewSocialPropagation(),
		appLog:       appLog,
	}
}

// SpreadGossip records an interaction for social propagation.
func (s *Service) SpreadGossip(ctx context.Context, fromNPC, toNPC *model.NPC, content string, locationID uint) {
	s.socialSpread.Spread(ctx, fromNPC, toNPC, content, locationID)
}

// HearGossip gets social context for an NPC at a location.
func (s *Service) HearGossip(npcID uint, locationID uint) string {
	return s.socialSpread.HearGossip(npcID, locationID, 3)
}

func (s *Service) FindInteractions(townID uint) ([]MatchPair, error) {
	npcs, err := s.npcRepo.FindByTownID(townID)
	if err != nil {
		return nil, err
	}

	byLocation := make(map[uint][]model.NPC)
	for _, npc := range npcs {
		byLocation[npc.LocationID] = append(byLocation[npc.LocationID], npc)
	}

	pairs := s.matcher.FindPairs(byLocation, s.relRepo)
	if len(pairs) > 3 {
		pairs = pairs[:3]
	}
	return pairs, nil
}

func (s *Service) GenerateInteraction(
	ctx context.Context,
	a, b *model.NPC,
	llmGen func(ctx context.Context, a, b *model.NPC) (*InteractionResult, error),
) *InteractionResult {
	tag := "default"
	rel, err := s.relRepo.FindBetween(a.ID, b.ID)
	if err == nil && rel != nil && rel.Tag != "" {
		tag = rel.Tag
	}

	if llmGen != nil && (a.Mood == "anxious" || a.Mood == "worried" || b.Mood == "anxious" || b.Mood == "worried") {
		result, err := llmGen(ctx, a, b)
		if err == nil && result != nil {
			s.appLog.Info("LLM generated interaction", "npc_a", a.Name, "npc_b", b.Name)
			return result
		}
	}

	tmpl := GetDialogueTemplate(tag)
	return &InteractionResult{
		Dialogue: []DialogueLine{
			{Speaker: a.Name, Speech: tmpl.Initiator, Emotion: tmpl.InitiatorEmotion},
			{Speaker: b.Name, Speech: tmpl.Responder, Emotion: tmpl.ResponderEmotion},
		},
		RelDeltas: []RelDelta{
			{FromNPCID: a.ID, ToNPCID: b.ID, Delta: 1, Reason: "casual conversation"},
			{FromNPCID: b.ID, ToNPCID: a.ID, Delta: 1, Reason: "casual conversation"},
		},
	}
}

func (s *Service) MarkMoved(npcID uint) {
	s.matcher.MarkMoved(npcID)
}

func (s *Service) MarkDone(npcID1, npcID2 uint) {
	s.matcher.MarkInteracted(npcID1, npcID2)
}
