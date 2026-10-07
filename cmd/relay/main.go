package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/relay"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("relay exited with error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig("../../internal/config") 

	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	db, err := database.NewPG(&cfg.DBCfg)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer db.Close()

	relayRepo := relay.NewRepository(db)
	producer := kafka.NewProducer(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.Topic)
	runner := relay.NewRunner(relayRepo, producer)
	slog.Info("relay started")
	runner.Run(ctx)
	defer producer.Close()
	return nil
}
