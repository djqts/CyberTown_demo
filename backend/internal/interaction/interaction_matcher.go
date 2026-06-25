package interaction

import (
	"sync"
	"time"

	"backend/internal/model"
	"backend/internal/repo"
)

type pairKey struct {
	a, b uint
}

type InteractionMatcher struct {
	mu              sync.Mutex
	lastInteraction map[pairKey]time.Time
	npcMovedAt      map[uint]time.Time // tracks when each NPC last moved
}

func NewInteractionMatcher() *InteractionMatcher {
	return &InteractionMatcher{
		lastInteraction: make(map[pairKey]time.Time),
		npcMovedAt:      make(map[uint]time.Time),
	}
}

// MarkMoved records that an NPC just arrived at a new location.
func (m *InteractionMatcher) MarkMoved(npcID uint) {
	m.mu.Lock()
	m.npcMovedAt[npcID] = time.Now()
	m.mu.Unlock()
}

type MatchPair struct {
	A        model.NPC
	B        model.NPC
	Priority int
}

func (m *InteractionMatcher) FindPairs(npcsByLocation map[uint][]model.NPC, relRepo *repo.RelationshipRepo) []MatchPair {
	var pairs []MatchPair

	for _, npcs := range npcsByLocation {
		if len(npcs) < 2 {
			continue
		}
		for i := 0; i < len(npcs); i++ {
			for j := i + 1; j < len(npcs); j++ {
				a, b := npcs[i], npcs[j]
				key := makePairKey(a.ID, b.ID)
				if last, ok := m.lastInteraction[key]; ok {
					if time.Since(last) < 2*time.Hour {
						continue
					}
				}

				// Prefer recently-moved NPCs but don't require it
				m.mu.Lock()
				aMoved := m.npcMovedAt[a.ID]
				bMoved := m.npcMovedAt[b.ID]
				m.mu.Unlock()
				recently := time.Now().Add(-90 * time.Second)
				if !aMoved.Before(recently) || !bMoved.Before(recently) {
					// At least one moved recently → bonus priority
				}

				priority := m.calcPriority(&a, &b, relRepo)
				pairs = append(pairs, MatchPair{A: a, B: b, Priority: priority})
			}
		}
	}

	sortPairs(pairs)
	return pairs
}

func (m *InteractionMatcher) MarkInteracted(npcID1, npcID2 uint) {
	key := makePairKey(npcID1, npcID2)
	m.mu.Lock()
	m.lastInteraction[key] = time.Now()
	m.mu.Unlock()
}

func (m *InteractionMatcher) calcPriority(a, b *model.NPC, relRepo *repo.RelationshipRepo) int {
	priority := 0
	// 关系值权重×2：有关系的NPC优先互动
	rel, err := relRepo.FindBetween(a.ID, b.ID)
	if err == nil && rel != nil {
		priority += int(rel.Affinity) * 2 // affinity 0-100 → 0-200 priority
	}
	// 同类型角色天然有共同话题
	if roleCategory(a.Role) == roleCategory(b.Role) {
		priority += 15
	}
	// 情绪共鸣：同为负面或正面情绪容易产生互动
	if isNegative(a.Mood) && isNegative(b.Mood) {
		priority += 10 // 同病相怜
	}
	if isPositive(a.Mood) && isPositive(b.Mood) {
		priority += 5 // 一起开心
	}
	return priority
}

func isNegative(mood string) bool {
	switch mood {
	case "anxious", "worried", "sad", "angry", "tired":
		return true
	}
	return false
}

func isPositive(mood string) bool {
	switch mood {
	case "cheerful", "happy", "excited", "inspired", "jolly", "playful":
		return true
	}
	return false
}

func makePairKey(a, b uint) pairKey {
	if a < b {
		return pairKey{a, b}
	}
	return pairKey{b, a}
}

func sortPairs(pairs []MatchPair) {
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Priority > pairs[i].Priority {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
}

func roleCategory(role string) string {
	switch role {
	case "铁匠", "木匠", "面包师":
		return "craft"
	case "咖啡馆主", "花店店主", "酒馆老板":
		return "merchant"
	case "镇长", "图书管理员", "教师", "医生":
		return "knowledge"
	case "农夫", "渔夫", "冒险者":
		return "outdoor"
	default:
		return "other"
	}
}
