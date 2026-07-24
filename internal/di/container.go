// internal/di/container.go

package di

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"mailForgeApi/internal/config"
	"mailForgeApi/internal/database"
	redisclient "mailForgeApi/internal/redisclient"
	"mailForgeApi/internal/routes"
	"mailForgeApi/internal/server"
	"mailForgeApi/pkg/logger"
	tokens "mailForgeApi/pkg/token"
)

func NewModules() fx.Option {
	return fx.Options(
		fx.Provide(config.NewInitConfig),
		fx.Provide(provideLogger),
		fx.Provide(database.NewDatabase),
		fx.Provide(redisclient.NewRedisClient),
		fx.Provide(providePrivateKey),
		fx.Provide(providePublicKey),
		fx.Provide(provideRefreshTokenManager),
		fx.Provide(routes.NewRouter),
		fx.Provide(server.NewServer),
		fx.Invoke(registerDBHooks),
		fx.Invoke(registerRedisHooks),
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
func provideRefreshTokenManager(client *redis.Client, cfg *config.Config) (tokens.RefreshTokenManager, error) {
	ttl, err := config.ParseExpiry(cfg.Jwt.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("parsing JWT_REFRESH_EXPIRY: %w", err)
	}
	return tokens.NewRefreshTokenManager(client, ttl), nil
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

func registerRedisHooks(
	lc fx.Lifecycle,
	client *redis.Client,
	log *logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis connection")
			return client.Close()
		},
	})
}
