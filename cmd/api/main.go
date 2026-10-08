package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/events"
	"webhook-relay/internal/subs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("api exited with error", "err", err)
		os.Exit(1)
	}	
}

func run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	cfg, err := config.LoadConfig("../../internal/config")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pg, err := database.NewPG(ctx, &cfg.DBCfg)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pg.Close()
	// if err = database.RunMigrations(pg.DB); err != nil {
	// 	return fmt.Errorf("run migrations: %w", err)
	// }

	mux := http.NewServeMux();

	// Subs
	subRepo := subs.NewRepo(pg)
	subServices := subs.NewService(subRepo)
	subRouters := subs.NewHandler(subServices)

	eventRepo := events.NewRepository(pg)
	eventService := events.NewEventService(eventRepo)
	eventHandler := events.NewHandler(eventService)


	subRouters.Register(mux)
	eventHandler.Register(mux)

	srv := &http.Server{
		Addr: ":8000",
		Handler: mux,
		ReadTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server started")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return 			
		}
		close(serverErr)
	}()

	select {
	case  <-ctx.Done():
	case srvErr := <-serverErr:
		if srvErr != nil {
			return fmt.Errorf("server failed: %w", srvErr)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "err", err)
	}
	return nil
}