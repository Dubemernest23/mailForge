package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.Logger
}

func New(env string) *Logger {
	var core zapcore.Core

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if env == "production" {
		// JSON logs for production (readable by log aggregators like Datadog, Loki)
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
	} else {
		// Colored console logs for development
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
	}

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{zap: zapLogger}
}

// Info logs a message at info level with optional fields
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Error logs a message at error level with optional fields
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Warn logs a message at warn level with optional fields
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Debug logs a message at debug level with optional fields
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Fatal logs then calls os.Exit(1)
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// With returns a child logger with pre-attached fields
// useful for attaching request_id, user_id etc to all logs in a context
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Sync flushes any buffered log entries — call this on app shutdown
func (l *Logger) Sync() {
	l.zap.Sync()
}
