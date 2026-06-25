package story

import "encoding/json"

type NPCEffect struct {
	NPCName string `json:"npc_name"`
	Mood    string `json:"mood,omitempty"`
	Goal    string `json:"goal,omitempty"`
}

type StoryEffects struct {
	NPCEffects []NPCEffect `json:"npc_effects"`
}

func ParseEffects(effectsJSON string) (*StoryEffects, error) {
	var effects StoryEffects
	if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
		return nil, err
	}
	return &effects, nil
}
