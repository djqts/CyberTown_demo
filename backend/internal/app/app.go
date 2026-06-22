package app

import (
	"context"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"

	"backend/internal/logger"
	"backend/internal/memory"
	"backend/internal/seed"
)

// Infra 聚合所有基础设施客户端。
type Infra struct {
	Redis  *redis.Client
	Qdrant *qdrant.Client
}

// InitMemory 初始化记忆系统，返回记忆服务。
func InitMemory(ctx context.Context, infra *Infra, appLog *logger.AppLogger) *memory.Service {
	short := memory.NewShortMemory(infra.Redis)
	long := memory.NewLongMemory(infra.Qdrant)

	// 初始化 Qdrant collections
	if err := memory.InitCollections(ctx, infra.Qdrant); err != nil {
		appLog.Error(err, "初始化 Qdrant collections 失败")
	} else {
		appLog.Info("Qdrant collections 已就绪")
	}

	// 导入世界知识
	seed.SeedWorldKnowledge(infra.Qdrant, appLog)

	return memory.NewService(short, long)
}
