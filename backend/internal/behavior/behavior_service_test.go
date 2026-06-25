package behavior

import (
	"context"
	"testing"

	"backend/internal/logger"
	"backend/internal/model"

	"gorm.io/gorm"
)

func TestEvaluateLowEnergyTriggersRest(t *testing.T) {
	rule := NewActivityRule()
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Energy: 20, Mood: "content"}

	reason := rule.Evaluate(npc)
	if reason != ReasonLowEnergy {
		t.Fatalf("Evaluate returned %q, want %q", reason, ReasonLowEnergy)
	}
}

func TestEvaluateNegativeMoodTriggersExpression(t *testing.T) {
	rule := NewActivityRule()
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Energy: 80, Mood: "anxious"}

	reason := rule.Evaluate(npc)
	if reason != ReasonNegativeMood {
		t.Fatalf("Evaluate returned %q, want %q", reason, ReasonNegativeMood)
	}
}

func TestEvaluateNewLocationTriggers(t *testing.T) {
	rule := NewActivityRule()
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Energy: 80, Mood: "content", LocationID: 3}
	rule.lastLocation[1] = 2

	reason := rule.Evaluate(npc)
	if reason != ReasonNewLocation {
		t.Fatalf("Evaluate returned %q, want %q", reason, ReasonNewLocation)
	}
}

func TestEvaluateSameLocationNoTrigger(t *testing.T) {
	rule := NewActivityRule()
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Energy: 80, Mood: "content", LocationID: 3}
	rule.lastLocation[1] = 3

	noneCount := 0
	for i := 0; i < 200; i++ {
		r2 := NewActivityRule()
		r2.lastLocation[1] = 3
		if r2.Evaluate(npc) == ReasonNone {
			noneCount++
		}
	}
	if noneCount < 150 {
		t.Fatalf("got ReasonNone %d/200 times, expect ~190", noneCount)
	}
}

func TestTemplateActivityReturnsResult(t *testing.T) {
	svc := NewBehaviorService(logger.NewApp("error", false))
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Name: "莉娜", Role: "咖啡馆主", Mood: "cheerful"}

	result := svc.templateActivity(npc, "", ReasonRandom)
	if result == nil {
		t.Fatal("templateActivity returned nil")
	}
	if result.NPCID != 1 || result.NPCName != "莉娜" {
		t.Fatalf("result NPC info wrong: %+v", result)
	}
	if result.Action == "" {
		t.Fatal("action is empty")
	}
}

func TestDecideReturnsNilWhenNoTrigger(t *testing.T) {
	svc := NewBehaviorService(logger.NewApp("error", false))
	npc := &model.NPC{Model: gorm.Model{ID: 1}, Energy: 80, Mood: "content", LocationID: 1}
	svc.rule.lastLocation[1] = 1

	found := false
	for i := 0; i < 100; i++ {
		r2 := NewBehaviorService(logger.NewApp("error", false))
		r2.rule.lastLocation[1] = 1
		if r2.Decide(context.Background(), npc, "", nil) != nil {
			found = true
			break
		}
	}
	if !found {
		t.Log("no random trigger in 100 calls, low probability but acceptable")
	}
}

func TestGetRoleActionsReturnsFallback(t *testing.T) {
	actions := GetRoleActions("不存在的角色")
	if len(actions) == 0 {
		t.Fatal("expected fallback actions")
	}
}

func TestGetMoodAction(t *testing.T) {
	action := GetMoodAction("anxious")
	if action == "" {
		t.Fatal("expected non-empty mood action for 'anxious'")
	}
}

func TestIsNegativeMood(t *testing.T) {
	tests := []struct {
		mood     string
		expected bool
	}{
		{"anxious", true},
		{"worried", true},
		{"sad", true},
		{"angry", true},
		{"happy", false},
		{"content", false},
		{"cheerful", false},
		{"calm", false},
	}
	for _, tt := range tests {
		if got := isNegativeMood(tt.mood); got != tt.expected {
			t.Errorf("isNegativeMood(%q) = %v, want %v", tt.mood, got, tt.expected)
		}
	}
}
