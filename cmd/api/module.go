package main

import (
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	"github.com/fercho/school-tracking/services/notification/pkg/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func AppModule() fx.Option {
	return fx.Options(
		env.Module,
		logger.Module,
		fx.Invoke(func(l *zap.Logger, cfg *env.Config) {
			l.Info("Starting notification service", zap.String("port", cfg.Port), zap.String("env", cfg.Environment))
		}),
	)
}
