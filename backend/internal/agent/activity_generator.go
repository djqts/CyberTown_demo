package agent

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/model"
)

// ActivityGenerator generates LLM-driven NPC activity descriptions.
type ActivityGenerator struct {
	runner *EinoRunner
}

func NewActivityGenerator(runner *EinoRunner) *ActivityGenerator {
	return &ActivityGenerator{runner: runner}
}

// Generate produces a personalized activity for the NPC based on their current state.
func (g *ActivityGenerator) Generate(ctx context.Context, npc *model.NPC, reason string) (string, error) {
	prompt := fmt.Sprintf(`你正在为游戏"晨曦镇"中的NPC生成主动行为描述。

NPC: %s (%s)
性格: %s
当前情绪: %s
当前位置: (建筑)
触发原因: %s

请用一句中文（不超过40字）描述该NPC正在做什么。只返回描述文字，不要加引号或解释。`, npc.Name, npc.Role, npc.Personality, npc.Mood, reason)

	resp, err := g.runner.Run(ctx, PromptInput{
		NPC:       npc,
		UserToken: "__activity__",
		UserInput: prompt,
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("empty response")
	}
	content := strings.TrimSpace(resp.Content)
	// Strip quotes
	content = strings.Trim(content, `"''`)
	if len(content) > 80 {
		content = content[:80]
	}
	return content, nil
}
