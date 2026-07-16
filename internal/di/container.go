// internal/di/container.go

package di

import (
	"context"
	"crypto/rsa"

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
		fx.Provide(providePrivateKey),
		fx.Provide(providePublicKey),
		fx.Provide(routes.NewRouter),
		fx.Provide(server.NewServer),
		fx.Invoke(registerDBHooks),
	)
}

func provideLogger(cfg *config.Config) *logger.Logger {
	return logger.New(cfg.Server.AppEnv)
}

// providePrivateKey loads and parses the RSA private key used to sign access tokens.
// Returning an error here (rather than panicking) means Fx surfaces key-loading
// failures as a clean app.Err() at boot, instead of a bare panic mid-startup.
func providePrivateKey(cfg *config.Config) (*rsa.PrivateKey, error) {
	return config.LoadPrivateKey(cfg.Jwt.PrivateKeyPath)
}

// providePublicKey loads and parses the RSA public key used to verify access tokens.
// Provided separately from the private key so modules that only need to verify
// tokens (e.g. middleware) never have the signing key in their dependency graph.
func providePublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	return config.LoadPublicKey(cfg.Jwt.PublicKeyPath)
}

func registerDBHooks(lc fx.Lifecycle, db *bun.DB, log *logger.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection", zap.String("status", "closing"))
			return db.Close()
		},
	})
}
