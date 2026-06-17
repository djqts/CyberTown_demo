package logger

import "log/slog"

// SensitiveString 实现 [slog.LogValuer]，防止敏感信息意外泄露到日志输出中。
// 无论底层字符串内容如何，日志中均显示为 "[REDACTED]"。
//
// 用法：
//
//	creds := logger.SensitiveString("sk-abc123")
//	log.Info("认证成功", "api_key", creds)
//	// 输出: ... "api_key": "[REDACTED]"
type SensitiveString string

// LogValue 返回 "[REDACTED]"，隐藏原始内容。
func (s SensitiveString) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

// MaskedJSON 实现 [slog.LogValuer]，用于需要字段级脱敏的结构化数据。
// 默认实现原样输出底层数据；嵌入或包装此类型并覆盖 [MaskedJSON.LogValue]
// 可实现自定义脱敏逻辑（如将 password 字段清零）。
//
// 用法：
//
//	type SafeUser struct { logger.MaskedJSON }
//	func (u SafeUser) LogValue() slog.Value {
//	    user := u.Data.(User)
//	    user.Password = "***"
//	    return slog.AnyValue(user)
//	}
type MaskedJSON struct {
	Data any
}

// LogValue 将包装的数据作为 slog 值原样输出。
// 嵌入类型可覆盖此方法实现自定义脱敏。
func (m MaskedJSON) LogValue() slog.Value {
	return slog.AnyValue(m.Data)
}
