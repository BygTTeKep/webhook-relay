package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/events"
	"webhook-relay/internal/logger"
	"webhook-relay/internal/subs"

	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	cfg, err := config.LoadConfig("../../internal/config")

	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %w", err)
		os.Exit(1)
	}
	log, err := logger.New("api", cfg.LoggerCfg.Dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %w", err)
		os.Exit(1)

	}
	code := 0
	if err := run(ctx, cfg, log); err != nil {
		log.Error("api exited with error", zap.Error(err))
		code = 1
	}
	stop()
	_ = log.Sync()
	os.Exit(code)
}

func run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {
	serverErr := make(chan error, 1)

	pg, err := database.NewPG(ctx, &cfg.DBCfg)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pg.Close()
	// if err = database.RunMigrations(pg.DB); err != nil {
	// 	return fmt.Errorf("run migrations: %w", err)
	// }

	mux := http.NewServeMux()

	// Subs
	subRepo := subs.NewRepo(pg)
	subServices := subs.NewService(subRepo)
	subRouters := subs.NewHandler(subServices, log)

	eventRepo := events.NewRepository(pg)
	eventService := events.NewEventService(eventRepo)
	eventHandler := events.NewHandler(eventService, log)

	subRouters.Register(mux)
	eventHandler.Register(mux)

	appPort := strconv.Itoa(cfg.AppCfg.Port)
	srv := &http.Server{
		Addr:              ":" + appPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("server started", zap.String("port", appPort))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
	case srvErr := <-serverErr:
		if srvErr != nil {
			return fmt.Errorf("server failed: %w", srvErr)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", zap.Error(err))
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
