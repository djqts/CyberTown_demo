package behavior

import (
	"math/rand"
	"time"

	"backend/internal/model"
)

// TriggerReason 表示行为触发的原因。
type TriggerReason string

const (
	ReasonLowEnergy    TriggerReason = "low_energy"
	ReasonNegativeMood TriggerReason = "negative_mood"
	ReasonRandom       TriggerReason = "random"
	ReasonNewLocation  TriggerReason = "new_location"
	ReasonNone         TriggerReason = ""
)

// ActivityRule 判断 NPC 是否需要生成主动行为。
type ActivityRule struct {
	lastLocation map[uint]uint // npcID -> last locationID
}

// NewActivityRule 创建行为规则实例。
func NewActivityRule() *ActivityRule {
	return &ActivityRule{
		lastLocation: make(map[uint]uint),
	}
}

// Evaluate 评估 NPC 是否应该触发主动行为，返回触发原因。
func (r *ActivityRule) Evaluate(npc *model.NPC) TriggerReason {
	if npc.Energy < 30 {
		return ReasonLowEnergy
	}

	if isNegativeMood(npc.Mood) {
		return ReasonNegativeMood
	}

	if rand.Intn(100) < 5 {
		return ReasonRandom
	}

	prevLoc, exists := r.lastLocation[npc.ID]
	r.lastLocation[npc.ID] = npc.LocationID
	if exists && prevLoc != npc.LocationID {
		return ReasonNewLocation
	}

	return ReasonNone
}

func isNegativeMood(mood string) bool {
	switch mood {
	case "anxious", "worried", "sad", "angry", "tired":
		return true
	}
	return false
}

// ShouldUseLLM 判断是否应调用 LLM 生成行为。
func ShouldUseLLM(moodChanged bool, storyActive bool, justInteracted bool) bool {
	return moodChanged || storyActive || justInteracted
}

// NowPtr 返回当前时间的指针。
func NowPtr() *time.Time {
	t := time.Now()
	return &t
}
