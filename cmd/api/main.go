package api

import (
	"context"
	"errors"
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


	cfg, err := config.LoadConfig("")
	if err != nil {
		stop()
	}
	pg, err := database.NewPG(&cfg.DBCfg)
	if err != nil {
		stop()
	}

	mux := http.NewServeMux();

	// Subs
	subRepo := subs.NewRepo(pg.DB)
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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "err", err)
	}
}