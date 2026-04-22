package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/migrate"

	"mailForgeApi/internal/config"
	migrationfiles "mailForgeApi/internal/migrations" // aliased to avoid collision
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|status]")
		os.Exit(1)
	}

	cfg := config.NewInitConfig()

	sqlDB, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	db := bun.NewDB(sqlDB, mysqldialect.New())

	m := migrate.NewMigrations()
	if err := m.Discover(migrationfiles.SQLMigrations); err != nil {
		fmt.Printf("failed to discover migrations: %v\n", err)
		os.Exit(1)
	}

	migrator := migrate.NewMigrator(db, m)
	ctx := context.Background()

	switch os.Args[1] {
	case "up":
		if err := migrator.Init(ctx); err != nil {
			fmt.Printf("init error: %v\n", err)
			os.Exit(1)
		}
		group, err := migrator.Migrate(ctx)
		if err != nil {
			fmt.Printf("migration error: %v\n", err)
			os.Exit(1)
		}
		if group.IsZero() {
			fmt.Println("nothing to migrate")
		} else {
			fmt.Printf("migrated: %s\n", group)
		}

	case "down":
		group, err := migrator.Rollback(ctx)
		if err != nil {
			fmt.Printf("rollback error: %v\n", err)
			os.Exit(1)
		}
		if group.IsZero() {
			fmt.Println("nothing to rollback")
		} else {
			fmt.Printf("rolled back: %s\n", group)
		}

	case "status":
		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			fmt.Printf("status error: %v\n", err)
			os.Exit(1)
		}
		for _, m := range ms {
			status := "pending"
			if !m.MigratedAt.IsZero() {
				status = "applied"
			}
			fmt.Printf("%-40s %s\n", m.Name, status)
		}

	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
