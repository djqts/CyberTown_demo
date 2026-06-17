# Logger

基于 Go 标准库 `log/slog` 的生产级日志封装，提供上下文感知日志、错误堆栈自动注入、敏感数据脱敏、采样和多输出支持。

## 特性

- **上下文富化** — 从 `context.Context` 自动提取 `trace_id`、`span_id` 等字段注入每条日志
- **错误自动堆栈** — `Error` 级别日志自动附加完整 goroutine 堆栈，且避免重复添加
- **敏感数据脱敏** — 通过 `slog.LogValuer` 接口实现 `SensitiveString`、`MaskedJSON` 等辅助类型
- **灵活配置** — 函数式选项模式，支持日志级别、格式、输出目标、源码位置、堆栈开关等
- **多输出** — 基于 `slog-multi` 的 Fanout 机制同时写文件和控制台
- **采样** — 集成 `slog-sampling` 统一采样器，高流量下降低日志量
- **本地优先** — 通过标准输出和本地文件解耦，便于对接 Vector / Fluentd 等日志收集工具

## 安装

```bash
go get github.com/samber/slog-multi
go get github.com/samber/slog-sampling
```

## 快速开始

```go
package main

import (
    "log/slog"
    "backed/internal/logger"
)

func main() {
    log := logger.New(
        logger.WithLevel(slog.LevelDebug),
        logger.WithFormat(logger.FormatJSON),
    )

    log.Info("server started", "port", 8080)
    log.Debug("config loaded", "path", "/etc/app.yml")

    if err := doWork(); err != nil {
        log.Error(err, "work failed", "component", "worker")
    }
}
```

## 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithLevel(lvl)` | 最低日志级别 | `slog.LevelInfo` |
| `WithFormat(f)` | 输出格式：`FormatJSON` / `FormatText` | `FormatJSON` |
| `WithWriter(w)` | 输出目标 | `os.Stdout` |
| `WithHandler(h)` | 直接设置 `slog.Handler`（优先于 Format/Writer） | — |
| `WithSource(bool)` | 是否记录源码位置 | `true` |
| `WithStackOnError(bool)` | Error 时自动附加堆栈 | `true` |
| `WithSampling(rate)` | 均匀采样率 (0.0 ~ 1.0) | 关闭 |
| `WithContextExtractor(fn)` | context → Attrs 提取函数 | — |
| `WithReplaceAttr(fn)` | 自定义属性替换（键名转换/脱敏） | — |
| `WithMultiWriter(handlers...)` | 额外输出 handler（Fanout） | — |
| `WithFileOutput(path)` | 便捷写文件（追加模式） | — |

## 高级用法

### OpenTelemetry 集成

```go
import "go.opentelemetry.io/otel/trace"

func otelExtractor(ctx context.Context) []slog.Attr {
    span := trace.SpanFromContext(ctx)
    if !span.IsRecording() {
        return nil
    }
    return []slog.Attr{
        slog.String("trace_id", span.SpanContext().TraceID().String()),
        slog.String("span_id",  span.SpanContext().SpanID().String()),
    }
}

log := logger.New(
    logger.WithContextExtractor(otelExtractor),
)
```

### 多输出（控制台 + 文件）

```go
f, _ := os.OpenFile("/var/log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

log := logger.New(
    logger.WithFormat(logger.FormatJSON),
    logger.WithMultiWriter(
        slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}),
        slog.NewJSONHandler(f, nil),
    ),
)
```

### 采样（高流量场景）

```go
log := logger.New(
    logger.WithSampling(0.1), // 保留约 10% 的日志
)
```

### 敏感数据脱敏

```go
log.Info("auth", "token", logger.SensitiveString("sk-abc123"))
// 输出: {"level":"INFO","msg":"auth","token":"[REDACTED]"}

// 自定义脱敏结构体
type SafeUser struct { logger.MaskedJSON }

func (u SafeUser) LogValue() slog.Value {
    user := u.Data.(User)
    user.Password = "***"
    return slog.AnyValue(user)
}
```

### 键名 Snake Case 转换

```go
import "strings"

func toSnake(s string) string {
    // ... snake_case 转换逻辑 ...
}

log := logger.New(
    logger.WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
        return slog.Attr{Key: toSnake(a.Key), Value: a.Value}
    }),
)
```

## 架构

```
New() 构造流程：

  cfg.handler  (用户提供)
       │
       ▼ (未提供时自动创建)
  baseHandler ─── slog.NewJSONHandler / slog.NewTextHandler
       │            (level, addSource, replaceAttr)
       ▼
  stacktraceMiddleware  ─── Error 级别注入 debug.Stack()
       │                      (已存在 stacktrace 时跳过)
       ▼
  contextHandler  ─── 从 context.Context 提取字段
       │                (trace_id, span_id, user_id ...)
       ▼
  slog-sampling  ─── 均匀采样器 (可选)
       │
       ▼
  slogmulti.Fanout  ─── 复制到多个 handler (可选)
       │
       ▼
  slog.New() → *Logger
```

中间件链的固定顺序保证了：
1. 堆栈在 Error 级别最早注入，确保被采样捕获
2. Context 字段在采样前注入，可按 trace_id 做调试采样
3. Fanout 在所有处理后分发，每条输出都包含完整信息

## API 参考

### Logger 方法

```go
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any)
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any)
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any)
func (l *Logger) ErrorContext(ctx context.Context, err error, msg string, args ...any)
func (l *Logger) Debug(msg string, args ...any)
func (l *Logger) Info(msg string, args ...any)
func (l *Logger) Warn(msg string, args ...any)
func (l *Logger) Error(err error, msg string, args ...any)
```

### 类型

```go
type Logger struct { *slog.Logger }    // 核心 Logger
type Config struct { ... }             // 配置（通过 Option 设置）
type Option func(*Config)              // 函数式选项
type ContextExtractor func(ctx context.Context) []slog.Attr  // Context 提取器
type SensitiveString string            // LogValuer：显示 [REDACTED]
type MaskedJSON struct { Data any }    // LogValuer：自定义脱敏基类
```

## 文件结构

```
logger/
├── logger.go       # Logger 类型、New() 构造函数、上下文便捷方法
├── config.go       # Config 结构体、Option 函数式选项
├── stacktrace.go   # Error 级别自动堆栈注入中间件
├── context.go      # Context 提取中间件
├── sensitive.go    # LogValuer 脱敏类型
├── logger_test.go  # 单元测试
└── README.md
```

## 最佳实践

- **生产环境**：配置日志轮转（如 `lumberjack`），避免磁盘写满
- **上下文抽取**：结合请求链注入 `request_id`、`user_id` 等业务字段
- **性能**：高并发下启用采样、必要时关闭 `WithStackOnError`
- **集中管理**：封装不绑定特定服务，通过 stdout 或文件对接 Vector / Fluentd
