package behavior

import (
	"math/rand"

	"backend/internal/agent"
	"backend/internal/model"
)

// EnumerateActions 根据NPC当前状态枚举合法候选动作
// 这是"有边界的自由"的核心——LLM只能从这些选项中选择
func EnumerateActions(npc *model.NPC, locationName string) []agent.CandidateAction {
	var candidates []agent.CandidateAction

	// 1. 工作动作（如果在自己建筑）
	if isHomeLocation(npc.Role, locationName) {
		workActions := RoleActivityPool[npc.Role]
		if len(workActions) > 0 {
			for i, act := range workActions {
				if i >= 3 {
					break // 最多3个工作选项
				}
				candidates = append(candidates, agent.CandidateAction{
					Action:     act,
					ActionType: "work",
					Priority:   80,
				})
			}
		}
	}

	// 2. 社交动作（根据外向性 EXTRAVERSION 调整选项）
	socialChance := 60 // 默认60%
	if npc.Name == "奥托" {
		socialChance = 20 // 低外向性
	}
	if npc.Name == "莉娜" || npc.Name == "米娅" {
		socialChance = 90 // 高外向性
	}
	if rand.Intn(100) < socialChance {
		candidates = append(candidates, agent.CandidateAction{
			Action:     "和附近的人聊几句，交流一下今天的见闻",
			ActionType: "talk",
			Priority:   50,
		})
	}

	// 3. 休息动作（如果精力低）
	if npc.Energy < 40 {
		candidates = append(candidates, agent.CandidateAction{
			Action: "找个安静的地方坐下休息一会儿，恢复精力", ActionType: "rest", Priority: 90,
		})
	}

	// 4. 探索动作（高开放性 NPC）
	if ocean, ok := agent.NPCOCEAN[npc.Name]; ok && ocean.Openness > 70 {
		candidates = append(candidates, agent.CandidateAction{
			Action: "在周围探索一下，看看有没有什么新鲜事", ActionType: "explore", Priority: 30,
		})
	}

	// 5. 地点相关的通用动作
	locActions := LocationActivityPool[locationName]
	for i, act := range locActions {
		if i >= 2 {
			break
		}
		candidates = append(candidates, agent.CandidateAction{
			Action:     act,
			ActionType: "idle",
			Priority:   40,
		})
	}

	// 6. 如果情绪负面，加入情绪表达动作
	if isNegativeMood(npc.Mood) {
		if moodAct := GetMoodAction(npc.Mood); moodAct != "" {
			candidates = append(candidates, agent.CandidateAction{
				Action: moodAct, ActionType: "idle", Priority: 85,
			})
		}
	}

	// 打乱顺序以避免位置偏差
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// 限制最多6个选项
	if len(candidates) > 6 {
		candidates = candidates[:6]
	}

	if len(candidates) == 0 {
		candidates = append(candidates, agent.CandidateAction{
			Action: "安静地待着，观察着周围的一切", ActionType: "idle", Priority: 20,
		})
	}

	return candidates
}

