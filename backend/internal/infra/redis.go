package infra

import (
	"context"
	"fmt"
	"time"

	config "backend/internal/config"
	"backend/internal/logger"

	"github.com/redis/go-redis/v9"
)

// RedisClient 包装 *redis.Client。
type RedisClient struct {
	*redis.Client
}

// NewRedisClient 创建并验证 Redis 连接。初始化过程通过 appLog 记录。
func NewRedisClient(appLog *logger.AppLogger) (*RedisClient, error) {
	cfg := config.AppConfig.Redis
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	appLog.Info("正在连接 Redis", "addr", addr, "db", 0)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		appLog.Error(err, "Redis 连接失败", "addr", addr)
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	appLog.Info("Redis 连接成功", "addr", addr)
	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
