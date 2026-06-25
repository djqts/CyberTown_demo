package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/model"
)

// OCEAN 人格维度 (0-100)
type OCEANTraits struct {
	Openness          int `json:"openness"`          // 开放性：好奇、创造性 vs 传统、保守
	Conscientiousness int `json:"conscientiousness"`  // 尽责性：自律、有条理 vs 随意、散漫
	Extraversion      int `json:"extraversion"`       // 外向性：社交、活跃 vs 独处、安静
	Agreeableness     int `json:"agreeableness"`      // 宜人性：合作、友善 vs 竞争、批判
	Neuroticism       int `json:"neuroticism"`        // 神经质：焦虑、敏感 vs 稳定、平和
}

// NPCOCEAN 预设人格（按角色设计）
var NPCOCEAN = map[string]OCEANTraits{
	"埃德蒙": {Openness: 50, Conscientiousness: 85, Extraversion: 70, Agreeableness: 80, Neuroticism: 30},
	"莉娜":   {Openness: 75, Conscientiousness: 70, Extraversion: 90, Agreeableness: 85, Neuroticism: 35},
	"艾琳":   {Openness: 80, Conscientiousness: 75, Extraversion: 25, Agreeableness: 70, Neuroticism: 40},
	"菲奥娜": {Openness: 85, Conscientiousness: 65, Extraversion: 75, Agreeableness: 80, Neuroticism: 30},
	"奥托":   {Openness: 30, Conscientiousness: 95, Extraversion: 20, Agreeableness: 45, Neuroticism: 55},
	"克莱尔": {Openness: 55, Conscientiousness: 90, Extraversion: 50, Agreeableness: 85, Neuroticism: 30},
	"杰克":   {Openness: 40, Conscientiousness: 80, Extraversion: 45, Agreeableness: 75, Neuroticism: 25},
	"沃尔特": {Openness: 35, Conscientiousness: 60, Extraversion: 15, Agreeableness: 55, Neuroticism: 20},
	"索菲亚": {Openness: 70, Conscientiousness: 80, Extraversion: 60, Agreeableness: 90, Neuroticism: 35},
	"皮埃尔": {Openness: 65, Conscientiousness: 75, Extraversion: 80, Agreeableness: 75, Neuroticism: 30},
	"玛莎":   {Openness: 50, Conscientiousness: 70, Extraversion: 85, Agreeableness: 70, Neuroticism: 30},
	"卢卡斯": {Openness: 95, Conscientiousness: 40, Extraversion: 55, Agreeableness: 65, Neuroticism: 60},
	"托马斯": {Openness: 30, Conscientiousness: 90, Extraversion: 25, Agreeableness: 60, Neuroticism: 25},
	"米娅":   {Openness: 90, Conscientiousness: 30, Extraversion: 95, Agreeableness: 90, Neuroticism: 50},
	"薇拉":   {Openness: 85, Conscientiousness: 60, Extraversion: 40, Agreeableness: 55, Neuroticism: 25},
}

// CandidateAction 候选动作——LLM从程序中枚举的合法动作中选择
type CandidateAction struct {
	Action      string `json:"action"`      // 人类可读的动作描述
	ActionType  string `json:"action_type"` // "move" | "talk" | "work" | "rest" | "explore"
	TargetID    uint   `json:"target_id,omitempty"`    // 目标位置或NPC的ID
	TargetName  string `json:"target_name,omitempty"`
	Priority    int    `json:"priority"`    // 程序计算的优先级 (0-100)
}

// AgentDecision NPC的最终决策
type AgentDecision struct {
	ChosenAction string `json:"chosen_action"`  // 选择的具体动作描述
	Reasoning    string `json:"reasoning"`      // 选择原因（内部独白）
	MoodChange   string `json:"mood_change,omitempty"` // 情绪变化
}

// NPCAgent NPC智能体：人格 + 记忆 → 有边界的自由决策
type NPCAgent struct {
	npc    *model.NPC
	ocean  OCEANTraits
	runner *EinoRunner
}

// NewNPCAgent creates an NPC Agent with personality.
func NewNPCAgent(npc *model.NPC, runner *EinoRunner) *NPCAgent {
	ocean := NPCOCEAN[npc.Name]
	if ocean.Openness == 0 {
		ocean = OCEANTraits{Openness: 50, Conscientiousness: 50, Extraversion: 50, Agreeableness: 50, Neuroticism: 50}
	}
	return &NPCAgent{npc: npc, ocean: ocean, runner: runner}
}

// GetName returns the NPC's name.
func (a *NPCAgent) GetName() string { return a.npc.Name }

// GetRole returns the NPC's role.
func (a *NPCAgent) GetRole() string { return a.npc.Role }

// GetOcean returns the personality traits.
func (a *NPCAgent) GetOcean() OCEANTraits { return a.ocean }

// DecideFromCandidates 从程序枚举的候选动作中，由LLM根据人格和记忆选择一个
// 这是"有边界的自由"核心实现——LLM只能在候选集中选择
func (a *NPCAgent) DecideFromCandidates(ctx context.Context, candidates []CandidateAction, memoryContext string) (*AgentDecision, error) {
	if len(candidates) == 0 {
		return &AgentDecision{ChosenAction: "安静地待着，观察周围", Reasoning: "没有合适的选项"}, nil
	}

	candidatesJSON, _ := json.Marshal(candidates)

	prompt := fmt.Sprintf(`你是晨曦镇的NPC。
姓名: %s | 职业: %s | 性格: %s
当前情绪: %s | 当前位置: (建筑内)
人格维度(0-100): O=%d C=%d E=%d A=%d N=%d
(O=开放性 C=尽责性 E=外向性 A=宜人性 N=神经质)

%s

## 候选动作（你只能从中选择一个）
%s

请以JSON格式输出你的决策，包含:
- chosen_action: 你选择的具体动作描述（必须来自候选列表）
- reasoning: 你为什么选择这个（基于你的性格和记忆，1句话）
- mood_change: 你的情绪是否有变化（可选，25种情绪之一）

{"chosen_action":"...","reasoning":"...","mood_change":"..."}`,
		a.npc.Name, a.npc.Role, truncate(a.npc.Personality, 60), a.npc.Mood,
		a.ocean.Openness, a.ocean.Conscientiousness, a.ocean.Extraversion, a.ocean.Agreeableness, a.ocean.Neuroticism,
		memoryContext, string(candidatesJSON))

	resp, err := a.runner.Run(ctx, PromptInput{NPC: a.npc, UserToken: "__decision__", UserInput: prompt})
	if err != nil || resp == nil {
		// Fallback: choose highest priority
		best := candidates[0]
		return &AgentDecision{ChosenAction: best.Action, Reasoning: "按优先级选择"}, nil
	}

	content := cleanJSON(resp.Content)
	var decision AgentDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		best := candidates[0]
		return &AgentDecision{ChosenAction: best.Action, Reasoning: "格式解析失败，默认选择"}, nil
	}
	if decision.ChosenAction == "" && len(candidates) > 0 {
		decision.ChosenAction = candidates[0].Action
	}

	return &decision, nil
}

// GeneratePlan 为NPC生成当日计划（每天开始时调用一次）
func (a *NPCAgent) GeneratePlan(ctx context.Context, timeOfDay string, memoryContext string) ([]string, error) {
	prompt := fmt.Sprintf(`你是晨曦镇的NPC %s（%s）。
性格: %s | 当前情绪: %s
人格: O=%d C=%d E=%d A=%d N=%d

现在是%s。请根据你的性格和记忆，列出今天接下来想做的3-5件事。

%s

以JSON数组格式输出: ["计划1", "计划2", "计划3"]`,
		a.npc.Name, a.npc.Role, truncate(a.npc.Personality, 60), a.npc.Mood,
		a.ocean.Openness, a.ocean.Conscientiousness, a.ocean.Extraversion, a.ocean.Agreeableness, a.ocean.Neuroticism,
		timeOfDay, memoryContext)

	resp, err := a.runner.Run(ctx, PromptInput{NPC: a.npc, UserToken: "__plan__", UserInput: prompt})
	if err != nil || resp == nil {
		return []string{fmt.Sprintf("继续%s", a.npc.CurrentGoal)}, nil
	}
	content := cleanJSON(resp.Content)
	var plans []string
	if err := json.Unmarshal([]byte(content), &plans); err != nil {
		return []string{fmt.Sprintf("继续%s", a.npc.CurrentGoal)}, nil
	}
	return plans, nil
}

// Reflect 每日反思——总结今天的重要事件，形成长期记忆
func (a *NPCAgent) Reflect(ctx context.Context, todaysEvents string) (string, error) {
	prompt := fmt.Sprintf(`你是晨曦镇的NPC %s（%s）。
今天发生了以下事件:
%s

请用1-2句话总结今天对你来说最重要的经历或感悟。只返回总结文字。`,
		a.npc.Name, a.npc.Role, todaysEvents)

	resp, err := a.runner.Run(ctx, PromptInput{NPC: a.npc, UserToken: "__reflect__", UserInput: prompt})
	if err != nil || resp == nil {
		return fmt.Sprintf("%s度过了平常的一天", a.npc.Name), nil
	}
	return strings.TrimSpace(resp.Content), nil
}
