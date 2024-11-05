package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestLogger_NewLogger(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantLvl zapcore.Level
	}{
		{
			name:    "default log level",
			level:   "",
			wantLvl: zapcore.InfoLevel,
		},
		{
			name:    "valid log level",
			level:   "DEBUG",
			wantLvl: zapcore.DebugLevel,
		},
		{
			name:    "invalid log level",
			level:   "INVALID",
			wantLvl: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := NewLogger(tt.level)
			assert.NoError(t, err)
			assert.NotNil(t, log)
			logLevel := log.Desugar().Core().Enabled(tt.wantLvl)
			assert.True(t, logLevel, "expected log level: %v", tt.wantLvl)
		})
	}
}
