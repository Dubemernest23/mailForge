package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"

	"mailForgeApi/internal/config"
	"mailForgeApi/pkg/logger"

	"go.uber.org/zap"
)

func NewDatabase(cfg *config.Config, log *logger.Logger) (*bun.DB, error) {

	// log.Info("database connected",
	// 	zap.String("host", cfg.DB.Host),
	// 	zap.String("port", fmt.Sprintf("%d", cfg.DB.Port)),
	// 	zap.String("name", cfg.DB.Name),
	// )
	sqlDB, err := sql.Open("mysql", cfg.Database.DSN)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// connection pool tuning
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	// verify the connection is actually alive
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to reach database: %w", err)
	}

	db := bun.NewDB(sqlDB, mysqldialect.New())

	// only attach query debugger in development
	if cfg.Server.AppEnv != "production" {
		db.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	log.Info("database connected",
		zap.String("host", cfg.DB.Host),
		zap.Int("max_open_conns", 25),
	)

	return db, nil
}
