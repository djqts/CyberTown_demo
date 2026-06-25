package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/model"
)

// InteractionGenerator generates LLM-driven NPC-to-NPC conversations.
type InteractionGenerator struct {
	runner *EinoRunner
}

func NewInteractionGenerator(runner *EinoRunner) *InteractionGenerator {
	return &InteractionGenerator{runner: runner}
}

type dialogueLine struct {
	Speaker string `json:"speaker"`
	Speech  string `json:"speech"`
	Action  string `json:"action"`
	Emotion string `json:"emotion"`
}

type interactionOutput struct {
	Dialogue    []dialogueLine `json:"dialogue"`
	MoodChanges map[string]string `json:"mood_changes"`
	RelDelta    int               `json:"rel_delta"`
}

// Generate produces a 2-4 turn conversation between two NPCs.
func (g *InteractionGenerator) Generate(ctx context.Context, npcA, npcB *model.NPC) ([]dialogueLine, map[string]string, int, error) {
	prompt := fmt.Sprintf(`你正在为游戏"晨曦镇"中的两个NPC生成自然对话。

NPC A: %s (%s), 性格: %s, 情绪: %s
NPC B: %s (%s), 性格: %s, 情绪: %s

请生成2-3轮简短中文对话（每人1-2句，不超过40字/句）。
输出JSON格式：
{"dialogue":[{"speaker":"名字","speech":"对话","action":"动作(可选,可为空)","emotion":"情绪"}],
 "mood_changes":{"名字":"新情绪(可选,无变化则跳过)"},
 "rel_delta":-1到3的关系变化值}`, npcA.Name, npcA.Role, truncate(npcA.Personality, 50), npcA.Mood,
		npcB.Name, npcB.Role, truncate(npcB.Personality, 50), npcB.Mood)

	resp, err := g.runner.Run(ctx, PromptInput{
		NPC:       npcA,
		UserToken: "__interaction__",
		UserInput: prompt,
	})
	if err != nil {
		return nil, nil, 0, err
	}
	if resp == nil {
		return nil, nil, 0, fmt.Errorf("empty response")
	}

	content := cleanJSON(resp.Content)
	var out interactionOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, nil, 0, fmt.Errorf("parse interaction JSON: %w", err)
	}
	if len(out.Dialogue) == 0 {
		return nil, nil, 0, fmt.Errorf("empty dialogue")
	}

	relDelta := out.RelDelta
	if relDelta < -3 {
		relDelta = -3
	}
	if relDelta > 3 {
		relDelta = 3
	}

	return out.Dialogue, out.MoodChanges, relDelta, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
