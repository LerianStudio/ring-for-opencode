package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{logger: l.logger.With(args...)}
}

type Options struct {
	Level   Level
	Output  io.Writer
	Verbose bool
}

type Option func(*Options)

func WithLevel(level Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

func WithOutput(w io.Writer) Option {
	return func(o *Options) {
		o.Output = w
	}
}

func WithVerbose(verbose bool) Option {
	return func(o *Options) {
		o.Verbose = verbose
		if verbose {
			o.Level = LevelDebug
		}
	}
}

func NewLogger(opts ...Option) Logger {
	options := &Options{
		Level:  LevelInfo,
		Output: os.Stderr,
	}
	for _, opt := range opts {
		opt(options)
	}

	handler := slog.NewTextHandler(options.Output, &slog.HandlerOptions{
		Level: options.Level,
	})

	return &slogLogger{logger: slog.New(handler)}
}

var (
	defaultLogger Logger = NewLogger()
	defaultMu     sync.RWMutex
)

func SetDefault(l Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger = l
}

func Default() Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultLogger
}

func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

func With(args ...any) Logger {
	return Default().With(args...)
}
