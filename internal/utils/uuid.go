package utils

import (
	"fmt"
	"time"
)

func GenerateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
