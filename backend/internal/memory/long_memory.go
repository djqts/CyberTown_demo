package memory

import (
	"context"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"
)

const (
	longMemoryCollection = "npc_memory"
)

// LongMemory Qdrant 长期记忆。
type LongMemory struct {
	client *qdrant.Client
}

// NewLongMemory 创建长期记忆存储。
func NewLongMemory(client *qdrant.Client) *LongMemory {
	return &LongMemory{client: client}
}

// Save 将重要对话写入 Qdrant。
func (m *LongMemory) Save(ctx context.Context, npcID uint, userToken, content string, embedding []float32) error {
	pointID := qdrant.NewIDNum(uint64(time.Now().UnixNano()))

	_, err := m.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: longMemoryCollection,
		Points: []*qdrant.PointStruct{
			{
				Id: pointID,
				Payload: map[string]*qdrant.Value{
					"npc_id":     {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(npcID)}},
					"user_token": {Kind: &qdrant.Value_StringValue{StringValue: userToken}},
					"content":    {Kind: &qdrant.Value_StringValue{StringValue: content}},
				},
				Vectors: qdrant.NewVectors(embedding...),
			},
		},
	})
	return err
}

// Search 检索与查询向量最相似的记忆。
func (m *LongMemory) Search(ctx context.Context, npcID uint, embedding []float32, limit int) ([]string, error) {
	result, err := m.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: longMemoryCollection,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchInt("npc_id", int64(npcID)),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var contents []string
	for _, pt := range result {
		if payload := pt.GetPayload(); payload != nil {
			if v, ok := payload["content"]; ok {
				contents = append(contents, v.GetStringValue())
			}
		}
	}
	return contents, nil
}

// SearchWorldKnowledge 检索世界知识（不限 NPC）。
func (m *LongMemory) SearchWorldKnowledge(ctx context.Context, embedding []float32, limit int) ([]string, error) {
	result, err := m.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "world_knowledge",
		Query:          qdrant.NewQuery(embedding...),
		Limit:          qdrant.PtrOf(uint64(limit)),
	})
	if err != nil {
		return nil, err
	}

	var contents []string
	for _, pt := range result {
		if payload := pt.GetPayload(); payload != nil {
			if v, ok := payload["content"]; ok {
				contents = append(contents, v.GetStringValue())
			}
		}
	}
	return contents, nil
}
