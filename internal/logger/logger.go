package logger

import (
	"go.uber.org/zap"
	"sync"
)

var logger *zap.Logger
var once sync.Once

// GetInstance returns a logger instance
func GetInstance() *zap.Logger {
	once.Do(func() {
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			zap.S().Fatalf("can't initialize zap logger: %v", err)
		}
	})
	return logger
}
