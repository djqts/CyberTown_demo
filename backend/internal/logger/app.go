// Package logger 提供 CyberTown 业务语义的日志功能。
// 通用封装（Logger、Config、Option 等）在本包其他文件；本文件专注业务级预配置和便捷助手。
package logger

import (
	"context"
	"log/slog"
)

// ---- 业务上下文键 ----

type ctxKey string

const (
	ctxKeyNPCID    ctxKey = "npc_id"    // NPC 唯一标识
	ctxKeyPlayerID ctxKey = "player_id" // 玩家唯一标识
	ctxKeyZoneID   ctxKey = "zone_id"   // 当前区域标识
	ctxKeyActionID ctxKey = "action_id" // 行为/任务标识
)

// ---- 预配置 ------

// NewAppLogger 创建适用于 CyberTown 的预配置 Logger。
// level: "debug" / "info" / "warn" / "error"
// isDev: 开发环境用 Text 格式 + 全量输出，生产环境用 JSON 格式 + 10% 采样。
func NewAppLogger(level string, isDev bool) *Logger {
	opts := []Option{
		WithStackOnError(true),
		WithContextExtractor(appContextExtractor),
		WithReplaceAttr(snakeCaseKeys),
	}

	if isDev {
		opts = append(opts,
			WithFormat(FormatText),
			WithSource(true),
		)
	} else {
		opts = append(opts,
			WithFormat(FormatJSON),
			WithSampling(0.1),
		)
	}

	switch level {
	case "debug":
		opts = append(opts, WithLevel(slog.LevelDebug))
	case "warn":
		opts = append(opts, WithLevel(slog.LevelWarn))
	case "error":
		opts = append(opts, WithLevel(slog.LevelError))
	default:
		opts = append(opts, WithLevel(slog.LevelInfo))
	}

	return New(opts...)
}

// ---- 业务语义 Logger ----

// AppLogger 包装 Logger，提供 CyberTown 业务域的便捷日志方法。
type AppLogger struct {
	*Logger
}

// NewApp 创建带业务语义的 AppLogger。
// level 同 NewAppLogger；isDev 控制输出格式和采样策略。
func NewApp(level string, isDev bool) *AppLogger {
	return &AppLogger{Logger: NewAppLogger(level, isDev)}
}

// NPC 记录 NPC 相关日志，自动携带 npc_id。
func (l *AppLogger) NPC(ctx context.Context, npcID, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyNPCID, npcID)
	l.InfoContext(ctx, msg, args...)
}

// NPCDebug 记录 NPC 调试日志。
func (l *AppLogger) NPCDebug(ctx context.Context, npcID, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyNPCID, npcID)
	l.DebugContext(ctx, msg, args...)
}

// NPCError 记录 NPC 错误日志，自动携带 npc_id 和堆栈。
func (l *AppLogger) NPCError(ctx context.Context, npcID string, err error, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyNPCID, npcID)
	l.ErrorContext(ctx, err, msg, args...)
}

// Player 记录玩家相关日志，自动携带 player_id。
func (l *AppLogger) Player(ctx context.Context, playerID, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyPlayerID, playerID)
	l.InfoContext(ctx, msg, args...)
}

// Zone 记录区域相关日志，自动携带 zone_id。
func (l *AppLogger) Zone(ctx context.Context, zoneID, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyZoneID, zoneID)
	l.InfoContext(ctx, msg, args...)
}

// Action 记录行为/任务日志，自动携带 action_id。
func (l *AppLogger) Action(ctx context.Context, actionID, msg string, args ...any) {
	ctx = context.WithValue(ctx, ctxKeyActionID, actionID)
	l.InfoContext(ctx, msg, args...)
}

// ---- 内部辅助 ----

// appContextExtractor 从 context 提取业务字段，注入每条日志。
// 提取的字段：npc_id、player_id、zone_id、action_id。
func appContextExtractor(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr
	if v := ctx.Value(ctxKeyNPCID); v != nil {
		attrs = append(attrs, slog.String("npc_id", v.(string)))
	}
	if v := ctx.Value(ctxKeyPlayerID); v != nil {
		attrs = append(attrs, slog.String("player_id", v.(string)))
	}
	if v := ctx.Value(ctxKeyZoneID); v != nil {
		attrs = append(attrs, slog.String("zone_id", v.(string)))
	}
	if v := ctx.Value(ctxKeyActionID); v != nil {
		attrs = append(attrs, slog.String("action_id", v.(string)))
	}
	return attrs
}

// snakeCaseKeys 保持 slog 默认的 snake_case 键名。
func snakeCaseKeys(groups []string, a slog.Attr) slog.Attr {
	return a
}
