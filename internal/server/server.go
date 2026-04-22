package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/fx"

	"mailForgeApi/internal/config"
)

func NewServer(cfg *config.Config, r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Server.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func StartServer(lc fx.Lifecycle, srv *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			fmt.Printf("[SERVER] Listening on http://localhost%s\n", srv.Addr)
			go srv.Serve(ln)
			go waitForShutdown(srv)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("[SERVER] Shutting down...")
			return srv.Shutdown(ctx)
		},
	})
}

func waitForShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("[SERVER] Signal received: %s\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("[SERVER] Forced shutdown: %v\n", err)
		os.Exit(1)
	}
}
