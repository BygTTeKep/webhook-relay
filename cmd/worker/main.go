package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/delivery"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/subs"
	"webhook-relay/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("worker exited with error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig("../../internal/config")
	slog.Info("config ", "kafka", cfg.KafkaCfg)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	db, err := database.NewPG(&cfg.DBCfg)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer db.Close()

	client := http.Client{} 

	subRepo := subs.NewRepo(db)
	deliveryRepo := delivery.NewRepository(db)
	deliveryService := delivery.NewService(deliveryRepo, &client)

	consumer:= kafka.NewConsumer(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.Topic) 
	defer consumer.Close()
	
	wr := worker.NewWorker(consumer)
	slog.Info("worker started")
	wr.Run(ctx, subRepo, deliveryService)
	return nil
}