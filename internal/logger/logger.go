package logger

import (
	"go.uber.org/zap"
)

func Init() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}