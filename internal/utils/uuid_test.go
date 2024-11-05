package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generate non-empty UUID",
		},
		{
			name: "generate unique UUIDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid1 := GenerateUUID()
			assert.NotEmpty(t, uuid1, "UUID should not be empty")
			time.Sleep(1 * time.Nanosecond)
			uuid2 := GenerateUUID()
			assert.NotEqual(t, uuid1, uuid2, "Two generated UUIDs should not be equal")
		})
	}
}
