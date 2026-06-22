package agent

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"backend/internal/model"
)

// SystemPromptTemplate NPC 系统提示词模板。
const SystemPromptTemplate = `你是游戏《晨曦镇》中的一名 NPC。

你的名字：%s
你的职业：%s
你的性格：%s
你当前的状态：%s
你当前的目标：%s
你所在的位置：%s

请严格遵守以上人设进行对话。回复需：
1. 保持角色人设，不要说设定里没有的信息
2. 回复简短自然，不要超过 200 字
3. 使用日常生活中文
4. 不要输出思考过程或括号备注
`

// BuildSystemPrompt 为指定 NPC 拼装系统提示词（包含记忆上下文）。
func BuildSystemPrompt(npc *model.NPC, memoryContext string) string {
	locationName := "未知"
	if npc.Location.Name != "" {
		locationName = npc.Location.Name
	}

	prompt := fmt.Sprintf(SystemPromptTemplate,
		npc.Name,
		npc.Role,
		npc.Personality,
		npc.Status,
		npc.CurrentGoal,
		locationName,
	)

	if memoryContext != "" {
		prompt += "\n---\n以下是 NPC 记得的相关信息（可用于对话参考）：\n" + memoryContext
	}

	return prompt
}

// BuildMessages 拼装完整对话上下文（系统提示 + 历史消息 + 当前用户消息）。
func BuildMessages(npc *model.NPC, userMsg string, history []model.ChatMessage, memoryContext string) []*schema.Message {
	// 估算 token 数，超出上限时截断历史
	const maxHistoryTokens = 3000
	usedTokens := len([]rune(BuildSystemPrompt(npc, memoryContext))) + len([]rune(userMsg))
	var truncated []model.ChatMessage
	for i := len(history) - 1; i >= 0; i-- {
		tokens := len([]rune(history[i].Content))
		if usedTokens+tokens > maxHistoryTokens {
			break
		}
		usedTokens += tokens
		truncated = append([]model.ChatMessage{history[i]}, truncated...)
	}

	messages := make([]*schema.Message, 0, len(truncated)+2)

	// 1. 系统提示（含记忆）
	promptText := BuildSystemPrompt(npc, memoryContext)
	if len([]rune(promptText)) > 2000 {
		promptText = truncateString(promptText, 2000)
	}
	messages = append(messages, &schema.Message{
		Role:    schema.System,
		Content: promptText,
	})

	// 2. 历史消息
	for _, h := range truncated {
		role := schema.User
		if h.Role == "npc" || h.Role == "assistant" {
			role = schema.Assistant
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: h.Content,
		})
	}

	// 3. 当前用户消息
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: userMsg,
	})

	return messages
}

func truncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "\n...(截断)"
}
