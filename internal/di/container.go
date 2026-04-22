package di

import (
	"context"

	"github.com/uptrace/bun"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"mailForgeApi/internal/config"
	"mailForgeApi/internal/database"
	"mailForgeApi/internal/routes"
	"mailForgeApi/internal/server"
	"mailForgeApi/pkg/logger"
)

func NewModules() fx.Option {
	return fx.Options(
		fx.Provide(config.NewInitConfig),
		fx.Provide(provideLogger),
		fx.Provide(database.NewDatabase),
		fx.Provide(routes.NewRouter),
		fx.Provide(server.NewServer),
		fx.Invoke(registerDBHooks),
	)
}

func provideLogger(cfg *config.Config) *logger.Logger {
	return logger.New(cfg.Server.AppEnv)
}

func registerDBHooks(lc fx.Lifecycle, db *bun.DB, log *logger.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection", zap.String("status", "closing"))
			return db.Close()
		},
	})
}
