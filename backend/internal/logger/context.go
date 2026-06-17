package logger

import (
	"context"
	"log/slog"
)

// contextHandler 是 [slog.Handler] 装饰器，通过用户提供的 [ContextExtractor]
// 函数从 [context.Context] 中提取属性，富化每条日志记录。
//
// 这是分布式追踪（OpenTelemetry、OpenCensus）、请求级元数据（user_id、tenant_id）
// 以及 context 中其他上下文信息的集成点。
//
// 提取器在堆栈中间件之后、采样之前运行，因此 trace_id 等上下文字段可影响采样决策。
type contextHandler struct {
	next    slog.Handler
	extract ContextExtractor
}

func newContextHandler(next slog.Handler, extract ContextExtractor) slog.Handler {
	return &contextHandler{next: next, extract: extract}
}

// Enabled 委托给链中的下一个 handler。
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle 在 ctx 上调用提取器，将所得属性追加到记录中，然后传递给下游。
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.extract != nil {
		attrs := h.extract(ctx)
		r.AddAttrs(attrs...)
	}
	return h.next.Handle(ctx, r)
}

// WithAttrs 将静态属性传递给下一个 handler，同时保留上下文提取行为。
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs), extract: h.extract}
}

// WithGroup 将分组名传递给下一个 handler，同时保留上下文提取行为。
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name), extract: h.extract}
}
