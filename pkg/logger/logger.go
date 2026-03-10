package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

var Module = fx.Provide(NewLogger)
