package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap's SugaredLogger to provide logging functionality with predefined configurations.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger creates a new Logger instance with the specified logging level.
func NewLogger(level string) (*Logger, error) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		log.Printf("Неверный уровень логирования: %s, используется уровень INFO по умолчанию", level)
		lvl = zapcore.InfoLevel
	}

	stdout := zapcore.AddSync(os.Stdout)
	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.TimeKey = "timestamp"
	developmentCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(developmentCfg),
		stdout,
		lvl,
	)

	return &Logger{zap.New(core).Sugar()}, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}
