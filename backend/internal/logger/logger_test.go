package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
)

// --- Helpers ---

func newBufLogger(opts ...Option) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	opts = append([]Option{WithWriter(&buf), WithSource(false)}, opts...)
	return New(opts...), &buf
}

func parseJSON(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse JSON log: %v\nraw: %s", err, buf.String())
	}
	return m
}

// --- Basic Logging ---

func TestNew_Defaults(t *testing.T) {
	log, buf := newBufLogger()
	log.Info("hello", "key", "val")

	m := parseJSON(t, buf)
	if m["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", m["level"])
	}
	if m["msg"] != "hello" {
		t.Errorf("expected msg 'hello', got %v", m["msg"])
	}
	if m["key"] != "val" {
		t.Errorf("expected key='val', got %v", m["key"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := New(WithFormat(FormatText), WithWriter(&buf), WithSource(false))
	log.Info("hello", "key", "val")

	out := buf.String()
	if !strings.Contains(out, "level=INFO") {
		t.Errorf("text format should contain 'level=INFO', got: %s", out)
	}
	if !strings.Contains(out, "msg=hello") {
		t.Errorf("text format should contain 'msg=hello', got: %s", out)
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	log := New(WithLevel(slog.LevelWarn), WithWriter(&buf), WithSource(false))

	buf.Reset()
	log.Debug("should be dropped")
	if buf.Len() != 0 {
		t.Error("Debug should be dropped when level is Warn")
	}

	buf.Reset()
	log.Info("should be dropped")
	if buf.Len() != 0 {
		t.Error("Info should be dropped when level is Warn")
	}

	buf.Reset()
	log.Warn("should appear")
	if buf.Len() == 0 {
		t.Error("Warn should appear when level is Warn")
	}

	buf.Reset()
	log.Error(io.EOF, "should appear")
	if buf.Len() == 0 {
		t.Error("Error should appear when level is Warn")
	}
}

// --- Context Methods ---

func TestDebugContext(t *testing.T) {
	log, buf := newBufLogger(WithLevel(slog.LevelDebug))
	log.DebugContext(context.Background(), "debug msg", "a", 1)
	m := parseJSON(t, buf)
	if m["msg"] != "debug msg" {
		t.Errorf("unexpected msg: %v", m["msg"])
	}
}

func TestInfoContext(t *testing.T) {
	log, buf := newBufLogger()
	log.InfoContext(context.Background(), "info msg")
	m := parseJSON(t, buf)
	if m["level"] != "INFO" {
		t.Errorf("unexpected level: %v", m["level"])
	}
}

func TestWarnContext(t *testing.T) {
	log, buf := newBufLogger()
	log.WarnContext(context.Background(), "warn msg")
	m := parseJSON(t, buf)
	if m["level"] != "WARN" {
		t.Errorf("unexpected level: %v", m["level"])
	}
}

func TestErrorContext(t *testing.T) {
	log, buf := newBufLogger()
	log.ErrorContext(context.Background(), io.EOF, "error msg", "detail", "x")

	m := parseJSON(t, buf)
	if m["level"] != "ERROR" {
		t.Errorf("unexpected level: %v", m["level"])
	}
	if m["msg"] != "error msg" {
		t.Errorf("unexpected msg: %v", m["msg"])
	}
	if m["detail"] != "x" {
		t.Errorf("unexpected detail: %v", m["detail"])
	}
	// Error field should be present
	errVal, ok := m["error"]
	if !ok {
		t.Error("expected 'error' key in error log")
	} else if errVal != "EOF" {
		t.Errorf("expected error='EOF', got %v", errVal)

	}
}

// --- Stacktrace ---

func TestStacktrace_OnError(t *testing.T) {
	log, buf := newBufLogger(WithStackOnError(true))
	log.Error(io.EOF, "something broke")

	m := parseJSON(t, buf)
	stack, ok := m["stacktrace"].(string)
	if !ok || stack == "" {
		t.Fatalf("expected non-empty stacktrace on Error, got: %v", m["stacktrace"])
	}
	if !strings.Contains(stack, "logger.TestStacktrace_OnError") {
		t.Errorf("stacktrace should contain the calling test function, got: %s", stack)
	}
}

func TestStacktrace_OnlyOnError(t *testing.T) {
	log, buf := newBufLogger(WithStackOnError(true))
	log.Info("normal message")

	m := parseJSON(t, buf)
	if _, exists := m["stacktrace"]; exists {
		t.Error("stacktrace should NOT appear on Info level")
	}
}

func TestStacktrace_Disabled(t *testing.T) {
	log, buf := newBufLogger(WithStackOnError(false))
	log.Error(io.EOF, "error without stack")

	m := parseJSON(t, buf)
	if _, exists := m["stacktrace"]; exists {
		t.Error("stacktrace should NOT appear when WithStackOnError is false")
	}
}

func TestStacktrace_NoDuplicates(t *testing.T) {
	log, buf := newBufLogger(WithStackOnError(true))
	// Manually add stacktrace before logging — middleware should NOT add another
	log.ErrorContext(context.Background(), io.EOF, "dup check", "stacktrace", "existing")

	m := parseJSON(t, buf)
	stack, ok := m["stacktrace"].(string)
	if !ok {
		t.Fatal("expected stacktrace to exist")
	}
	if strings.Count(stack, "goroutine") > 1 {
		t.Error("stacktrace should not be duplicated when already present")
	}
}

// Convenience without context

func TestConvenienceMethods(t *testing.T) {
	log, buf := newBufLogger(WithLevel(slog.LevelDebug))

	buf.Reset()
	log.Debug("debug")
	if buf.Len() == 0 {
		t.Error("Debug should log")
	}

	buf.Reset()
	log.Info("info")
	if buf.Len() == 0 {
		t.Error("Info should log")
	}

	buf.Reset()
	log.Warn("warn")
	if buf.Len() == 0 {
		t.Error("Warn should log")
	}

	buf.Reset()
	log.Error(io.EOF, "error")
	m := parseJSON(t, buf)
	if m["level"] != "ERROR" {
		t.Error("Error should log at ERROR level")
	}
}

// --- Context Extraction ---

func TestContextExtractor(t *testing.T) {
	extractor := func(ctx context.Context) []slog.Attr {
		traceID, _ := ctx.Value("trace_id").(string)
		return []slog.Attr{slog.String("trace_id", traceID)}
	}

	log, buf := newBufLogger(WithContextExtractor(extractor))

	ctx := context.WithValue(context.Background(), "trace_id", "abc123")
	log.InfoContext(ctx, "request processed")

	m := parseJSON(t, buf)
	if m["trace_id"] != "abc123" {
		t.Errorf("expected trace_id=abc123, got %v", m["trace_id"])
	}
}

func TestContextExtractor_NotConfigured(t *testing.T) {
	log, buf := newBufLogger()
	ctx := context.WithValue(context.Background(), "trace_id", "abc123")
	log.InfoContext(ctx, "no extractor")

	m := parseJSON(t, buf)
	if _, exists := m["trace_id"]; exists {
		t.Error("trace_id should NOT appear when no extractor is configured")
	}
}

// --- Sensitive Data ---

func TestSensitiveString(t *testing.T) {
	log, buf := newBufLogger()
	log.Info("auth", "token", SensitiveString("sk-secret-123"))

	m := parseJSON(t, buf)
	if m["token"] != "[REDACTED]" {
		t.Errorf("expected token=[REDACTED], got %v", m["token"])
	}
}

func TestMaskedJSON(t *testing.T) {
	log, buf := newBufLogger()
	log.Info("user", "data", MaskedJSON{Data: map[string]any{"name": "alice"}})

	m := parseJSON(t, buf)
	data, ok := m["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data to be an object, got %T", m["data"])
	}
	if data["name"] != "alice" {
		t.Errorf("expected name=alice, got %v", data["name"])
	}
}

// --- MultiWriter ---

func TestMultiWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	h1 := slog.NewJSONHandler(&buf1, nil)
	h2 := slog.NewJSONHandler(&buf2, nil)

	log := New(WithMultiWriter(h1, h2), WithSource(false))
	log.Info("fanout test")

	for i, buf := range []*bytes.Buffer{&buf1, &buf2} {
		if buf.Len() == 0 {
			t.Errorf("handler %d should receive log output", i+1)
		}
		m := parseJSON(t, buf)
		if m["msg"] != "fanout test" {
			t.Errorf("handler %d: expected msg='fanout test', got %v", i+1, m["msg"])
		}
	}
}

// --- WithHandler ---

func TestWithHandler_CustomHandler(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})

	log := New(WithHandler(h))
	log.Info("should be dropped")
	if buf.Len() != 0 {
		t.Error("Info should be dropped when custom handler has Error level")
	}

	buf.Reset()
	log.Error(io.EOF, "should appear")
	if buf.Len() == 0 {
		t.Error("Error should appear with custom handler")
	}
}

// --- ReplaceAttr ---

func TestReplaceAttr(t *testing.T) {
	log, buf := newBufLogger(WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "secret" {
			return slog.Attr{} // drop
		}
		return a
	}))

	log.Info("data", "secret", "hidden", "public", "visible")
	m := parseJSON(t, buf)

	if _, exists := m["secret"]; exists {
		t.Error("secret key should be dropped by ReplaceAttr")
	}
	if m["public"] != "visible" {
		t.Errorf("public should remain visible, got %v", m["public"])
	}
}

// --- Sampling ---

func TestSampling_Rate1_KeepsAll(t *testing.T) {
	log, buf := newBufLogger(WithSampling(1.0))
	for i := range 50 {
		log.Info("sampled", "i", i)
	}
	// At rate 1.0, every record should pass through.
	// Count JSON objects (each line is one JSON record).
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 50 {
		t.Errorf("rate 1.0: expected 50 lines, got %d", len(lines))
	}
}

func TestSampling_Rate0_DropsAll(t *testing.T) {
	log, buf := newBufLogger(WithSampling(0.0))
	for range 100 {
		log.Info("sampled")
	}
	if buf.Len() != 0 {
		t.Errorf("rate 0.0: expected no output, got %d bytes", buf.Len())
	}
}

// --- Edge Cases ---

func TestEmptyArgs(t *testing.T) {
	log, buf := newBufLogger()
	log.Info("no args")
	m := parseJSON(t, buf)
	if m["msg"] != "no args" {
		t.Errorf("unexpected msg: %v", m["msg"])
	}
}

func TestNilError(t *testing.T) {
	log, buf := newBufLogger()
	log.Error(nil, "nil error")
	m := parseJSON(t, buf)
	errVal, ok := m["error"]
	if !ok {
		t.Error("expected 'error' key when err is nil")
	}
	if errVal != nil {
		t.Errorf("expected null error value, got %v", errVal)
	}
}

func TestSourceDisabled(t *testing.T) {
	var buf bytes.Buffer
	log := New(WithWriter(&buf), WithSource(false))
	log.Info("no source")
	m := parseJSON(t, &buf)
	if _, exists := m["source"]; exists {
		t.Error("source should NOT appear when WithSource is false")
	}
}

// Test that the non-context Error convenience method still registers correctly.
func TestError_WithoutContext(t *testing.T) {
	log, buf := newBufLogger()
	log.Error(io.EOF, "test error", "k", "v")
	m := parseJSON(t, buf)
	if m["level"] != "ERROR" {
		t.Errorf("expected ERROR, got %v", m["level"])
	}
	if m["k"] != "v" {
		t.Errorf("expected k=v, got %v", m["k"])
	}
}
