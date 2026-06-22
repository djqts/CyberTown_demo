package websocket

// MessageType 定义 WebSocket 消息类型。
const (
	MsgTypeUserMessage = "user.message"
	MsgTypeNPCReplied  = "npc.replied"
)

// Message WebSocket 推送消息结构。
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// UserMessage 用户发给 NPC 的消息。
type UserMessage struct {
	NPCID     uint   `json:"npc_id"`
	Content   string `json:"content"`
	UserToken string `json:"user_token"`
}

// NPCReplied NPC 回复消息。
type NPCReplied struct {
	NPCID     uint   `json:"npc_id"`
	NPCName   string `json:"npc_name"`
	Content   string `json:"content"`
	UserToken string `json:"user_token"`
}
