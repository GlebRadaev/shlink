package logger

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func resetOnce() {
	once = sync.Once{}
}

func TestLogger_NewLogger_Level(t *testing.T) {
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
		resetOnce()
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.wantLvl, logger.Level())
		})
	}
}

func TestLogger_LoadLogger(t *testing.T) {
	type args struct {
		level zapcore.Level
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "default log level",
			args:    args{level: zapcore.InfoLevel},
			wantErr: nil,
		},
		{
			name:    "valid log level",
			args:    args{level: zapcore.DebugLevel},
			wantErr: nil,
		},
		{
			name:    "invalid log level",
			args:    args{level: zapcore.FatalLevel},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadLogger(tt.args.level)
			assert.NotNil(t, logger)
			assert.Equal(t, tt.args.level, logger.Level())
		})
	}
}
