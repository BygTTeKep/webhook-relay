package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/delivery"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/logger"
	"webhook-relay/internal/subs"
	"webhook-relay/internal/worker"

	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	cfg, err := config.LoadConfig("../../internal/config")
	if err != nil {
		fmt.Errorf("load config: %w", err)
		os.Exit(1)
	}
	log, err := logger.New("worker", cfg.LoggerCfg.Dev)
	if err != nil {
		fmt.Errorf("init logger: %w", err)
		os.Exit(1)
	}
	code := 0
	if err := run(ctx, cfg, log); err != nil {
		log.Error("worker exited with error", zap.Error(err))
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

	client := http.Client{}

	subRepo := subs.NewRepo(db)
	deliveryRepo := delivery.NewRepository(db)
	deliveryService := delivery.NewService(deliveryRepo, &client)

	consumer := kafka.NewConsumer(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.Topic)
	defer consumer.Close()

	wr := worker.NewWorker(consumer, log)
	log.Info("worker started")
	wr.Run(ctx, subRepo, deliveryService)
	return nil
}
