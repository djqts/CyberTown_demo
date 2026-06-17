// Package logger 基于 log/slog 的生产级日志封装。
//
// 提供上下文感知日志、错误自动堆栈注入、敏感数据脱敏（LogValuer）、
// 采样、多输出（控制台 + 文件）等功能，通过函数式选项 API 配置。
//
// 快速开始：
//
//	log := logger.New(
//	    logger.WithFormat(logger.FormatJSON),
//	    logger.WithLevel(slog.LevelDebug),
//	)
//	log.Info("服务启动", "port", 8080)
//
// 中间件链按固定顺序组装以保证正确性：
//
//	基础 handler → 堆栈中间件 → 上下文提取器 → 采样 → 扇出
//
// 此顺序确保上下文字段对采样决策可见，且堆栈在记录到达下游 handler 之前注入。
package logger

import (
	"context"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
)

// Logger 包装 [*slog.Logger]，增加接收 context.Context 的便捷方法。
// 对 Error 级别的方法，自动附加 error 值和堆栈跟踪（启用 [WithStackOnError] 时）。
//
// 所有日志方法均支持并发安全调用。
type Logger struct {
	*slog.Logger
	config Config
}

// New 根据给定的 [Option] 函数创建 Logger，并组装 handler 中间件链。
//
// 默认配置（可通过选项覆盖）：
//   - 级别：[slog.LevelInfo]
//   - 格式：[FormatJSON]
//   - AddSource：true
//   - StackOnError：true
//   - Writer：[os.Stdout]
//
// 中间件链顺序：
//  1. 基础 handler（JSON/Text，配置级别、源码位置和属性替换）
//  2. 堆栈中间件（Error 时注入 debug.Stack，启用时）
//  3. 上下文提取器（从 context.Context 注入字段，配置时）
//  4. 采样中间件（均匀限速，启用时）
//  5. 扇出（使用 [WithMultiWriter] 时复制到多个 handler）
func New(opts ...Option) *Logger {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	var handler slog.Handler
	baseHandler := cfg.handler

	// 优先使用用户提供的 handler，否则根据格式/输出设置构建。
	if baseHandler == nil {
		w := cfg.writer
		if w == nil {
			w = os.Stdout
		}
		ho := &slog.HandlerOptions{
			Level:       cfg.level,
			AddSource:   cfg.addSource,
			ReplaceAttr: cfg.replaceAttr,
		}
		switch cfg.format {
		case FormatJSON:
			baseHandler = slog.NewJSONHandler(w, ho)
		default:
			baseHandler = slog.NewTextHandler(w, ho)
		}
	}

	// 堆栈注入必须在采样之前，确保每个错误都被捕获。
	if cfg.stackOnError {
		baseHandler = newStacktraceMiddleware(baseHandler)
	}

	// 上下文富化在采样之前运行，让 trace_id 等字段影响采样决策。
	if cfg.contextExtractor != nil {
		baseHandler = newContextHandler(baseHandler, cfg.contextExtractor)
	}

	// 采样包装整个管线，按配置的比例丢弃记录。
	if cfg.sampling.enabled {
		baseHandler = slogmulti.
			Pipe(cfg.sampling.option.NewMiddleware()).
			Handler(baseHandler)
	}

	// 扇出将每条记录复制到所有已配置的 handler（如同时写控制台和文件）。
	if len(cfg.multiWriters) > 0 {
		handlers := []slog.Handler{baseHandler}
		for _, w := range cfg.multiWriters {
			handlers = append(handlers, w)
		}
		handler = slogmulti.Fanout(handlers...)
	} else {
		handler = baseHandler
	}

	return &Logger{
		Logger: slog.New(handler),
		config: cfg,
	}
}

// DebugContext 以 Debug 级别记录日志，从 ctx 中提取字段（通过已配置的 [ContextExtractor]）。
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.DebugContext(ctx, msg, args...)
}

// InfoContext 以 Info 级别记录日志，从 ctx 中提取字段（通过已配置的 [ContextExtractor]）。
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}

// WarnContext 以 Warn 级别记录日志，从 ctx 中提取字段（通过已配置的 [ContextExtractor]）。
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext 以 Error 级别记录日志，自动将 error 值附加到 "error" 键下。
// 启用 [WithStackOnError] 时，额外在 "stacktrace" 键下附加完整堆栈。
// ctx 中的字段会优先提取。
func (l *Logger) ErrorContext(ctx context.Context, err error, msg string, args ...any) {
	attrs := make([]any, 0, len(args)+1)
	attrs = append(attrs, slog.Any("error", err))
	attrs = append(attrs, args...)
	l.Logger.ErrorContext(ctx, msg, attrs...)
}

// Debug 以 Debug 级别记录日志，使用 background context。
func (l *Logger) Debug(msg string, args ...any) {
	l.DebugContext(context.Background(), msg, args...)
}

// Info 以 Info 级别记录日志，使用 background context。
func (l *Logger) Info(msg string, args ...any) {
	l.InfoContext(context.Background(), msg, args...)
}

// Warn 以 Warn 级别记录日志，使用 background context。
func (l *Logger) Warn(msg string, args ...any) {
	l.WarnContext(context.Background(), msg, args...)
}

// Error 以 Error 级别记录日志，使用 background context。
// 自动附加 error 和堆栈（启用时）。
func (l *Logger) Error(err error, msg string, args ...any) {
	l.ErrorContext(context.Background(), err, msg, args...)
}
