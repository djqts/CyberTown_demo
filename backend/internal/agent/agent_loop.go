package agent

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/model"
)

// AgentLoop 封装 NPC 的"感知-规划-执行"完整决策循环。
// 将原本分散在多个 Worker 中的决策逻辑统一为一个闭环。
type AgentLoop struct {
	agent  *NPCAgent
	memSvc MemoryRecall // 记忆召回接口
}

// MemoryRecall 记忆召回接口（避免循环导入 memory 包）
type MemoryRecall interface {
	Recall(ctx context.Context, npcID uint, userToken, currentMsg string) string
}

// NewAgentLoop 创建Agent决策循环。
func NewAgentLoop(npc *model.NPC, runner *EinoRunner, memSvc MemoryRecall) *AgentLoop {
	return &AgentLoop{
		agent:  NewNPCAgent(npc, runner),
		memSvc: memSvc,
	}
}

// PerceiveContext 感知阶段：收集当前情境的所有信息。
// 返回结构化的上下文，供后续规划和决策使用。
func (a *AgentLoop) PerceiveContext(ctx context.Context, locationName string, nearbyNPCs []string, currentEvents []string) string {
	var sb strings.Builder

	// 1. 基础状态
	sb.WriteString(fmt.Sprintf("【当前状态】\n位置: %s | 情绪: %s | 精力: %d | 目标: %s\n",
		locationName, a.agent.npc.Mood, a.agent.npc.Energy, a.agent.npc.CurrentGoal))

	// 2. 人格特征
	o := a.agent.ocean
	sb.WriteString(fmt.Sprintf("\n【人格】O=%d C=%d E=%d A=%d N=%d\n", o.Openness, o.Conscientiousness, o.Extraversion, o.Agreeableness, o.Neuroticism))

	// 3. 记忆召回（RAG）
	memories := a.memSvc.Recall(ctx, a.agent.npc.ID, "__perceive__", locationName)
	if memories != "" {
		sb.WriteString(fmt.Sprintf("\n【相关记忆】\n%s\n", truncateString(memories, 300)))
	}

	// 4. 环境感知
	if len(nearbyNPCs) > 0 {
		sb.WriteString(fmt.Sprintf("\n【附近NPC】%s\n", strings.Join(nearbyNPCs, "、")))
	}

	// 5. 当前事件
	if len(currentEvents) > 0 {
		sb.WriteString(fmt.Sprintf("\n【当前事件】%s\n", strings.Join(currentEvents, "；")))
	}

	return sb.String()
}

// GenerateIntent 规划阶段：基于感知的上下文，生成一个高层意图。
// 返回简单的动作描述即可，具体执行由引擎处理。
func (a *AgentLoop) GenerateIntent(ctx context.Context, perception string) string {
	prompt := fmt.Sprintf(`你是晨曦镇的NPC，请根据以下感知信息，生成一个简短的动作意图（1句话，10字以内，无需解释）。

%s`, perception)

	resp, err := a.agent.runner.Run(ctx, PromptInput{
		NPC: a.agent.npc, UserToken: "__intent__", UserInput: prompt,
	})
	if err != nil || resp == nil {
		return a.agent.npc.CurrentGoal
	}
	return strings.TrimSpace(resp.Content)
}

// ExecuteAction 执行阶段：将意图转化为具体可广播的动作描述。
// npcLocationName 用于位置感知的动作选择。
func (a *AgentLoop) ExecuteAction(ctx context.Context, intent string, perception string) string {
	prompt := fmt.Sprintf(`你是晨曦镇的NPC %s（%s）。

你的意图是：%s

请将意图扩展为一句完整的动作描述（15-40字，中文），描述你此刻正在做什么。
只返回描述文字。`, a.agent.npc.Name, a.agent.npc.Role, intent)

	resp, err := a.agent.runner.Run(ctx, PromptInput{
		NPC: a.agent.npc, UserToken: "__action__", UserInput: prompt,
	})
	if err != nil || resp == nil {
		return intent
	}
	return strings.TrimSpace(resp.Content)
}

func truncateString(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
