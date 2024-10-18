package logger

import (
	"log"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	logger *zap.SugaredLogger
)

func loadLogger(level zapcore.Level) {
	stdout := zapcore.AddSync(os.Stdout)

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.TimeKey = "timestamp"
	developmentCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	developmentConsoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	core := zapcore.NewTee(
		zapcore.NewCore(developmentConsoleEncoder, stdout, level),
	)
	logger = zap.New(core).Sugar()
}

func NewLogger(level ...string) *zap.SugaredLogger {
	once.Do(func() {
		lvl := zapcore.InfoLevel
		if len(level) > 0 {
			if err := lvl.UnmarshalText([]byte(level[0])); err != nil {
				log.Printf("Invalid log level: %s, defaulting to INFO", level[0])
			}
		}
		loadLogger(lvl)
	})
	return logger
}
func SyncLogger() {
	if logger != nil {
		_ = logger.Sync()
	}
}
