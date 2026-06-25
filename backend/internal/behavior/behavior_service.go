package behavior

import (
	"context"

	"backend/internal/logger"
	"backend/internal/model"
)

// ActivityResult 行为生成结果。
type ActivityResult struct {
	NPCID   uint   `json:"npc_id"`
	NPCName string `json:"npc_name"`
	Action  string `json:"action"`
	Mood    string `json:"mood"`
	Thought string `json:"thought,omitempty"`
}

// BehaviorService 管理 NPC 主动行为决策。
type BehaviorService struct {
	rule   *ActivityRule
	appLog *logger.AppLogger
}

// NewBehaviorService 创建行为服务。
func NewBehaviorService(appLog *logger.AppLogger) *BehaviorService {
	return &BehaviorService{
		rule:   NewActivityRule(),
		appLog: appLog,
	}
}

// Decide 对单个 NPC 进行行为决策，返回生成的 ActivityResult 或 nil。
func (s *BehaviorService) Decide(
	ctx context.Context,
	npc *model.NPC,
	locationName string,
	llmGen func(ctx context.Context, npc *model.NPC, reason TriggerReason) (*ActivityResult, error),
) *ActivityResult {
	reason := s.rule.Evaluate(npc)
	if reason == ReasonNone {
		return nil
	}

	if llmGen != nil {
		result, err := llmGen(ctx, npc, reason)
		if err == nil && result != nil {
			s.appLog.Info("LLM generated activity", "npc_id", npc.ID, "reason", reason)
			return result
		}
		s.appLog.Warn("LLM activity fallback to template", "npc_id", npc.ID, "error", err)
	}

	return s.templateActivity(npc, locationName, reason)
}

func (s *BehaviorService) templateActivity(npc *model.NPC, locationName string, reason TriggerReason) *ActivityResult {
	var action string

	switch reason {
	case ReasonLowEnergy:
		action = "疲惫地坐下休息，恢复精力"
	case ReasonNegativeMood:
		if moodAction := GetMoodAction(npc.Mood); moodAction != "" {
			action = moodAction
		} else {
			action = "情绪不太好，一个人静静待着"
		}
	case ReasonNewLocation, ReasonRandom:
		action = GetActivity(npc.Role, locationName)
	default:
		action = GetActivity(npc.Role, locationName)
	}

	return &ActivityResult{
		NPCID:   npc.ID,
		NPCName: npc.Name,
		Action:  action,
		Mood:    npc.Mood,
	}
}
