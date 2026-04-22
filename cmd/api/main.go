package main

import (
	"log"
	"mailForgeApi/internal/di"
	"mailForgeApi/internal/server"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// func main() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("[WARN] no .env file found, using system environment")
// 	}

// 	fx.New(
// 		di.NewModules(),
// 		fx.Invoke(server.StartServer),
// 	).Run()
// }

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
