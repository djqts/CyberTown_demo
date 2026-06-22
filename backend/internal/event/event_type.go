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
	ActorTypeSystem = "system"
	ActorTypeNPC    = "npc"
	ActorTypeUser   = "user"
)
