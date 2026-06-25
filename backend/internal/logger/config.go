package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	slogsampling "github.com/samber/slog-sampling"
)

// [WithFormat] 的格式常量。
const (
	FormatText = "text" // 人类可读的 key=value 输出。
	FormatJSON = "json" // 机器可解析的 JSON 输出。
)

// Config 包含所有日志配置。字段为未导出，强制通过 [Option] 函数设置，
// 保证前向兼容性。
type Config struct {
	level        slog.Leveler
	format       string
	writer       io.Writer
	handler      slog.Handler
	addSource    bool
	stackOnError bool

	sampling struct {
		enabled bool
		option  slogsampling.UniformSamplingOption
	}

	contextExtractor func(ctx context.Context) []slog.Attr
	replaceAttr      func(groups []string, a slog.Attr) slog.Attr

	multiWriters []slog.Handler
}

// defaultConfig 返回应用选项前的基准配置。
func defaultConfig() Config {
	return Config{
		level:        slog.LevelInfo,
		format:       FormatJSON,
		addSource:    true,
		stackOnError: true,
	}
}

// Option 是 [New] 的函数式选项类型。
type Option func(*Config)

// WithLevel 设置最低日志级别。低于此级别的记录将被丢弃。
//
//	logger.New(logger.WithLevel(slog.LevelDebug))
func WithLevel(level slog.Leveler) Option {
	return func(c *Config) { c.level = level }
}

// WithFormat 设置输出格式。可选 [FormatJSON] 或 [FormatText]。
//
//	logger.New(logger.WithFormat(logger.FormatText))
func WithFormat(f string) Option {
	return func(c *Config) { c.format = f }
}

// WithWriter 设置输出目标，默认为 [os.Stdout]。
//
//	f, _ := os.Create("/var/log/app.log")
//	logger.New(logger.WithWriter(f))
func WithWriter(w io.Writer) Option {
	return func(c *Config) { c.writer = w }
}

// WithHandler 直接设置自定义 [slog.Handler]。设置后，WithFormat 和 WithWriter
// 将被忽略——调用者自行负责 handler 的级别、格式和目标。
func WithHandler(h slog.Handler) Option {
	return func(c *Config) { c.handler = h }
}

// WithSource 启用或禁用每条日志记录的源码位置（文件:行号）。默认为 true。
func WithSource(b bool) Option {
	return func(c *Config) { c.addSource = b }
}

// WithStackOnError 启用或禁用 Error 级别日志自动注入堆栈跟踪。
// 堆栈通过 [runtime/debug.Stack] 捕获，存储在 "stacktrace" 键下。默认为 true。
//
// 如果生产环境中每条错误都捕获完整堆栈的性能开销不可接受，请关闭此选项。
func WithStackOnError(b bool) Option {
	return func(c *Config) { c.stackOnError = b }
}

// WithSampling 启用均匀随机采样，rate 范围 0.0 ~ 1.0。
// rate=0.1 表示约 10% 的日志记录被保留。
// 适用于高吞吐量服务，记录每条日志代价过高时。
//
//	logger.New(logger.WithSampling(0.1)) // 保留约 10% 记录
func WithSampling(rate float64) Option {
	return func(c *Config) {
		c.sampling.enabled = true
		c.sampling.option = slogsampling.UniformSamplingOption{
			Rate: rate,
		}
	}
}

// ContextExtractor 是从 [context.Context] 中提取 [slog.Attr] 字段的函数。
// 每条日志记录时调用，用于注入分布式追踪标识、用户信息、请求 ID 等。
//
// OpenTelemetry 示例：
//
//	func otelExtractor(ctx context.Context) []slog.Attr {
//	    span := trace.SpanFromContext(ctx)
//	    return []slog.Attr{
//	        slog.String("trace_id", span.SpanContext().TraceID().String()),
//	        slog.String("span_id",  span.SpanContext().SpanID().String()),
//	    }
//	}
type ContextExtractor func(ctx context.Context) []slog.Attr

// WithContextExtractor 注册一个 [ContextExtractor]，每条日志记录时从 context 中提取字段。
//
//	logger.New(logger.WithContextExtractor(otelExtractor))
func WithContextExtractor(fn ContextExtractor) Option {
	return func(c *Config) {
		c.contextExtractor = fn
	}
}

// WithReplaceAttr 设置在属性写入前对其转换或丢弃的回调函数。
// 常见用途：键名规范化（CamelCase → snake_case）、脱敏已知敏感键、添加全局字段。
//
//	logger.New(logger.WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
//	    if a.Key == "password" {
//	        return slog.Attr{} // 丢弃该字段
//	    }
//	    return a
//	}))
func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(c *Config) {
		c.replaceAttr = fn
	}
}

// WithMultiWriter 添加额外的 [slog.Handler] 实例，每条日志记录都会分发到它们（扇出）。
// 用于同时写入控制台和文件，或发送日志到外部服务。
//
//	logger.New(
//	    logger.WithMultiWriter(
//	        slog.NewJSONHandler(os.Stdout, nil),
//	        slog.NewJSONHandler(logFile, nil),
//	    ),
//	)
func WithMultiWriter(handlers ...slog.Handler) Option {
	return func(c *Config) {
		c.multiWriters = append(c.multiWriters, handlers...)
	}
}

// WithFileOutput 是便捷方法，打开（或创建）指定文件路径并设置为日志输出目标。
// 文件以追加模式打开。
//
//	logger.New(logger.WithFileOutput("/var/log/app.log"))
func WithFileOutput(path string) Option {
	return func(c *Config) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("无法打开日志文件: " + err.Error())
		}
		c.writer = f
	}
}
