// internal/di/auth.go
package di

import (
	"crypto/rsa"
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/fx"

	"mailForgeApi/internal/config"
	"mailForgeApi/internal/modules/auth"
	tokens "mailForgeApi/pkg/token"
)

// AuthModule bundles every auth-specific provider — repo, service, handler,
// and the access-token expiry duration that only the auth service needs.
func AuthModule() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				provideAccessExpiry,
				fx.ResultTags(`name:"accessExpiry"`),
			),
		),
		fx.Provide(provideAuthRepo),
		fx.Provide(
			fx.Annotate(
				provideAuthService,
				fx.ParamTags(``, ``, ``, `name:"accessExpiry"`),
			),
		),
		fx.Provide(provideAuthHandler),
	)
}

func provideAccessExpiry(cfg *config.Config) (time.Duration, error) {
	return config.ParseExpiry(cfg.Jwt.AccessExpiry)
}

func provideAuthRepo(db *bun.DB) auth.AuthRepo {
	return auth.NewAuthRepository(db)
}

func provideAuthService(
	repo auth.AuthRepo,
	refreshMgr tokens.RefreshTokenManager,
	privateKey *rsa.PrivateKey,
	accessExpiry time.Duration,
) *auth.Service {
	return auth.NewService(repo, refreshMgr, privateKey, accessExpiry)
}

func provideAuthHandler(svc *auth.Service) *auth.Handler {
	return auth.NewHandler(svc)
}
