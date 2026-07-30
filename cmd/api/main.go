package main

import (
	"log"
	"mailForgeApi/internal/di"
	"mailForgeApi/internal/server"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// @title MailForge API
// @version 1.0.0
// @description MailForge backend API. Phase B covers authentication and identity — registration, login, token refresh, and logout.
// @host localhost:3010
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT access token.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] no .env file found, using system environment")
	}

	app := fx.New(
		di.NewModules(),
		fx.Invoke(server.StartServer),
	)

	app.Run()

	if app.Err() != nil {
		log.Fatalf("[FATAL] %v", app.Err())
		os.Exit(1)
	}

	os.Exit(0)
}
