package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ramisoul84/resreview-server/internal/config"
	"github.com/rs/zerolog"
)

var (
	globalLogger Logger
	once         sync.Once
)

type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	WithFields(fields map[string]any) Logger
}

type logger struct {
	zerolog.Logger
}

func New(cfg *config.Config) Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	level := parseLevel(cfg.Logging.Level)
	output := buildOutput(cfg)

	zlog := zerolog.New(output).
		Level(level).
		With().
		Str("service", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("env", cfg.App.Env).
		Timestamp().
		Logger()

	return &logger{Logger: zlog}
}

func InitGlobal(cfg *config.Config) {
	once.Do(func() {
		globalLogger = New(cfg)
	})
}

func Get() Logger {
	if globalLogger == nil {
		panic("logger: Get() called before InitGlobal()")
	}
	return globalLogger
}

// ─── Package-level shortcuts ──────────────────────────────────────────────────
func Debug() *zerolog.Event { return Get().Debug() }
func Info() *zerolog.Event  { return Get().Info() }
func Warn() *zerolog.Event  { return Get().Warn() }
func Error() *zerolog.Event { return Get().Error() }
func Fatal() *zerolog.Event { return Get().Fatal() }

func (l *logger) WithFields(fields map[string]any) Logger {
	ctx := l.Logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	return &logger{Logger: ctx.Logger()}
}

func buildOutput(cfg *config.Config) io.Writer {
	if cfg.Logging.Output == "file" {
		return buildFileOutput(cfg.Logging.File)
	}

	if cfg.IsDevelopment() {
		return zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC822,
		}
	}
	return os.Stdout
}

func buildFileOutput(path string) io.Writer {
	// Ensure parent directory exists before trying to open the file.
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Printf("logger: failed to create log directory %q: %v — falling back to stdout\n", dir, err)
			return os.Stdout
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Printf("logger: failed to open log file %q: %v — falling back to stdout\n", path, err)
		return os.Stdout
	}
	return file
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
