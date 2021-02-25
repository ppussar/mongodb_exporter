package logger

import (
	"go.uber.org/zap"
	"log"
	"sync"
)

var Logger *zap.Logger
var once sync.Once

func GetInstance() *zap.Logger {
	once.Do(func() {
		var err error
		Logger, err = zap.NewProduction()
		if err != nil {
			log.Fatalf("can't initialize zap logger: %v", err)
		}
	})
	return Logger
}
