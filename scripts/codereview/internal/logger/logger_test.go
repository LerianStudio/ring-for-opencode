package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		l := NewLogger()
		require.NotNil(t, l)
	})

	t.Run("with custom output", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l := NewLogger(WithOutput(buf))

		l.Info("test message", "key", "value")

		output := buf.String()
		assert.Contains(t, output, "test message")
		assert.Contains(t, output, "key=value")
	})

	t.Run("with verbose mode", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l := NewLogger(WithOutput(buf), WithVerbose(true))

		l.Debug("debug message")

		assert.Contains(t, buf.String(), "debug message")
	})

	t.Run("debug filtered at info level", func(t *testing.T) {
		buf := &bytes.Buffer{}
		l := NewLogger(WithOutput(buf), WithLevel(LevelInfo))

		l.Debug("should not appear")

		assert.Empty(t, buf.String())
	})
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		logFn  func(Logger)
		expect bool
	}{
		{"debug at debug level", LevelDebug, func(l Logger) { l.Debug("msg") }, true},
		{"info at debug level", LevelDebug, func(l Logger) { l.Info("msg") }, true},
		{"warn at debug level", LevelDebug, func(l Logger) { l.Warn("msg") }, true},
		{"error at debug level", LevelDebug, func(l Logger) { l.Error("msg") }, true},
		{"debug at info level", LevelInfo, func(l Logger) { l.Debug("msg") }, false},
		{"info at info level", LevelInfo, func(l Logger) { l.Info("msg") }, true},
		{"debug at warn level", LevelWarn, func(l Logger) { l.Debug("msg") }, false},
		{"info at warn level", LevelWarn, func(l Logger) { l.Info("msg") }, false},
		{"warn at warn level", LevelWarn, func(l Logger) { l.Warn("msg") }, true},
		{"debug at error level", LevelError, func(l Logger) { l.Debug("msg") }, false},
		{"error at error level", LevelError, func(l Logger) { l.Error("msg") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			l := NewLogger(WithOutput(buf), WithLevel(tt.level))
			tt.logFn(l)

			if tt.expect {
				assert.NotEmpty(t, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	l := NewLogger(WithOutput(buf))

	child := l.With("component", "test")
	child.Info("message")

	output := buf.String()
	assert.Contains(t, output, "component=test")
	assert.Contains(t, output, "message")
}

func TestDefaultLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	original := Default()
	defer SetDefault(original)

	SetDefault(NewLogger(WithOutput(buf)))

	Info("test info", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "test info")
	assert.Contains(t, output, "key=value")
}

func TestPackageLevelFunctions(t *testing.T) {
	buf := &bytes.Buffer{}
	original := Default()
	defer SetDefault(original)

	SetDefault(NewLogger(WithOutput(buf), WithLevel(LevelDebug)))

	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 4)
}

func TestStructuredLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	l := NewLogger(WithOutput(buf))

	l.Info("Starting phase", "name", "build", "timeout", 30)
	l.Error("Phase failed", "error", "connection timeout", "duration", 5.5)

	output := buf.String()
	assert.Contains(t, output, "name=build")
	assert.Contains(t, output, "timeout=30")
	assert.Contains(t, output, "error=\"connection timeout\"")
	assert.Contains(t, output, "duration=5.5")
}
