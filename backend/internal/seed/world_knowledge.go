package seed

import (
	"context"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"

	"backend/internal/logger"
	"backend/internal/memory"
)

var worldFacts = []string{
	"晨曦镇是一个宁静的小镇，居民友善和睦。镇上有广场、咖啡馆、钟楼等地点。",
	"莉娜是咖啡馆的咖啡师，热情开朗，喜欢和每位客人闲聊，每天为客人冲泡今日特调。",
	"奥托是钟楼里的钟表匠，沉默寡言，对机械有极致追求，负责校准钟楼齿轮。",
	"米娅是镇上唯一的邮差，好奇心旺盛，喜欢打探镇上的新鲜事，每天派送信件。",
	"广场是晨曦镇的中心，居民们经常在这里聚会和交流。",
	"咖啡馆是莉娜工作的地方，也是镇上最受欢迎的社交场所。",
	"钟楼是晨曦镇最高的建筑，奥托每天都在这里修理和校准古老的钟表。",
	"镇上的邮局每天早晨分拣信件，米娅负责将所有信件准时送达。",
	"晨曦镇每周五晚上在广场举办集市，所有居民都会参加。",
	"咖啡馆的今日特调是莉娜每天早上决定的，配方秘而不宣。",
}

// SeedWorldKnowledge 将世界知识向量化后写入 Qdrant（幂等，collection 不存在时跳过）。
func SeedWorldKnowledge(client *qdrant.Client, appLog *logger.AppLogger) {
	ctx := context.Background()

	exists, err := client.CollectionExists(ctx, "world_knowledge")
	if err != nil || !exists {
		appLog.Warn("world_knowledge collection 不存在，跳过世界知识初始化", "err", err)
		return
	}

	// 检查是否已有数据
	result, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "world_knowledge",
		Query:          qdrant.NewQuery(make([]float32, memory.EmbedDim)...),
		Limit:          qdrant.PtrOf(uint64(1)),
	})
	if err == nil && len(result) > 0 {
		appLog.Info("世界知识已存在，跳过")
		return
	}

	appLog.Info("开始导入世界知识", "count", len(worldFacts))
	for i, fact := range worldFacts {
		vec := memory.Embed(fact)
		pointID := qdrant.NewIDNum(uint64(i + 1))

		_, err := client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: "world_knowledge",
			Points: []*qdrant.PointStruct{
				{
					Id: pointID,
					Payload: map[string]*qdrant.Value{
						"content": {Kind: &qdrant.Value_StringValue{StringValue: fact}},
					},
					Vectors: qdrant.NewVectors(vec...),
				},
			},
		})
		if err != nil {
			appLog.Error(err, "导入世界知识失败", "index", i)
			return
		}
		time.Sleep(50 * time.Millisecond) // 避免 Qdrant 过载
	}
	appLog.Info("世界知识导入完成", "count", len(worldFacts))
}
