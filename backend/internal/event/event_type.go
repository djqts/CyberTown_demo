package event

const (
	EventTypeTownTick       = "town.tick"
	EventTypeNPCMoveRequest = "npc.move.requested"
	EventTypeNPCMoved       = "npc.moved"
	EventTypeBroadcast      = "town.event.broadcast"
)

const (
	EventTypeUserMessageSent = "user.message.sent"
	EventTypeNPCReplied      = "npc.replied"
)

const (
	EventTypeNPCIdleAction = "npc.idle.action"
)

// Day 9: NPC 主动行为
const (
	EventTypeNPCActivityRequired  = "npc.activity.required"
	EventTypeNPCActivityGenerated = "npc.activity.generated"
	EventTypeNPCThoughtGenerated  = "npc.thought.generated"
	EventTypeNPCMoodChanged       = "npc.mood.changed"
)

// Day 10: NPC 互动
const (
	EventTypeNPCInteractionRequired  = "npc.interaction.required"
	EventTypeNPCInteractionGenerated = "npc.interaction.generated"
	EventTypeNPCRelationshipChanged  = "npc.relationship.changed"
)

// Day 11: 故事事件
const (
	EventTypeStoryEventTriggered = "story.event.triggered"
	EventTypeStoryEffectApplied  = "story.effect.applied"
	EventTypeNPCGoalChanged      = "npc.goal.changed"
	EventTypeTownNewsGenerated   = "town.news.generated"
)

const (
	ActorTypeSystem = "system"
	ActorTypeNPC    = "npc"
	ActorTypeUser   = "user"
)
