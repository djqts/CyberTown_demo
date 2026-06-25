package interaction

type DialogueLine struct {
	Speaker string `json:"speaker"`
	Speech  string `json:"speech"`
	Action  string `json:"action"`
	Emotion string `json:"emotion"`
}

type InteractionResult struct {
	Dialogue    []DialogueLine  `json:"dialogue"`
	MoodChanges map[uint]string `json:"mood_changes"`
	RelDeltas   []RelDelta      `json:"rel_deltas"`
}

type RelDelta struct {
	FromNPCID uint   `json:"from_npc_id"`
	ToNPCID   uint   `json:"to_npc_id"`
	Delta     int    `json:"delta"`
	Reason    string `json:"reason"`
}
