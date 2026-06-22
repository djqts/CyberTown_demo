package infra

import (
	"context"
	"fmt"
	"time"

	qdrant "github.com/qdrant/go-client/qdrant"

	config "backend/internal/config"
	"backend/internal/logger"
)

// QdrantClient 包装 *qdrant.Client。
type QdrantClient struct {
	*qdrant.Client
}

// NewQdrantClient 创建 Qdrant 客户端并验证连接。
func NewQdrantClient(appLog *logger.AppLogger) (*QdrantClient, error) {
	cfg := config.AppConfig.Qdrant

	appLog.Info("正在连接 Qdrant", "host", cfg.Host, "port", cfg.Port)

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	})
	if err != nil {
		appLog.Error(err, "Qdrant 连接失败", "host", cfg.Host, "port", cfg.Port)
		return nil, fmt.Errorf("qdrant new client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.HealthCheck(ctx); err != nil {
		client.Close()
		appLog.Error(err, "Qdrant 健康检查失败", "host", cfg.Host)
		return nil, fmt.Errorf("qdrant health check: %w", err)
	}

	appLog.Info("Qdrant 连接成功", "host", cfg.Host)
	return &QdrantClient{Client: client}, nil
}
