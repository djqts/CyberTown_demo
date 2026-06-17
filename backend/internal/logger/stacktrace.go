package logger

import (
	"context"
	"log/slog"
	"runtime/debug"
)

// stacktraceMiddleware 是 [slog.Handler] 装饰器，向每条 Error 级别记录中
// 注入完整的 goroutine 堆栈跟踪。
//
// 防护规则：
//   - 仅对 [slog.LevelError] 生效。
//   - 当记录中已存在 "stacktrace" 键时跳过（防止经过多个 handler 时重复叠加）。
//
// 性能说明：[runtime/debug.Stack] 捕获完整 goroutine 堆栈，涉及遍历调用栈。
// 对于生产环境中错误率高的路径，建议通过 [WithStackOnError](false) 关闭。
type stacktraceMiddleware struct {
	next slog.Handler
}

func newStacktraceMiddleware(next slog.Handler) slog.Handler {
	return &stacktraceMiddleware{next: next}
}

// Enabled 委托给链中的下一个 handler。
func (h *stacktraceMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle 检查记录级别。若为 Error，则在尚未存在的情况下追加 "stacktrace" 字段，
// 然后将记录传递给下游。
func (h *stacktraceMiddleware) Handle(ctx context.Context, r slog.Record) error {
	if r.Level == slog.LevelError {
		hasStack := false
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "stacktrace" {
				hasStack = true
				return false // 提前终止迭代
			}
			return true
		})
		if !hasStack {
			r.AddAttrs(slog.String("stacktrace", string(debug.Stack())))
		}
	}
	return h.next.Handle(ctx, r)
}

// WithAttrs 将静态属性传递给下一个 handler，同时保留堆栈行为。
func (h *stacktraceMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &stacktraceMiddleware{next: h.next.WithAttrs(attrs)}
}

// WithGroup 将分组名传递给下一个 handler，同时保留堆栈行为。
func (h *stacktraceMiddleware) WithGroup(name string) slog.Handler {
	return &stacktraceMiddleware{next: h.next.WithGroup(name)}
}
