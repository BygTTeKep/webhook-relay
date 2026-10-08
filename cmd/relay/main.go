package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/logger"
	"webhook-relay/internal/relay"

	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	cfg, err := config.LoadConfig("../../internal/config")

	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %w", err)
		os.Exit(1)
	}
	log, err := logger.New("relay", cfg.LoggerCfg.Dev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %w", err)
		os.Exit(1)
	}
	code := 0
	if err := run(ctx, cfg, log); err != nil {
		log.Error("relay exited with error", zap.Error(err))
		code = 1
	}
	stop()
	_ = log.Sync()
	os.Exit(code)
}

func run(ctx context.Context, cfg *config.Config, log *zap.Logger) error {

	db, err := database.NewPG(ctx, &cfg.DBCfg)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer db.Close()

	relayRepo := relay.NewRepository(db)
	producer := kafka.NewProducer(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.Topic)
	runner := relay.NewRunner(relayRepo, producer, log)
	log.Info("relay started")
	runner.Run(ctx)
	defer producer.Close()
	return nil
}
