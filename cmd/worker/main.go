package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	statsv1 "webhook-relay/cmd/gen/stats/v1"
	"webhook-relay/internal/config"
	"webhook-relay/internal/database"
	"webhook-relay/internal/delivery"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/logger"
	"webhook-relay/internal/subs"
	"webhook-relay/internal/worker"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	statsConn, err := grpc.NewClient(cfg.GrpcServer.StatsServ, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc stats conn: %w", err)
	}
	defer statsConn.Close()

	statsClient := statsv1.NewStatsServiceClient(statsConn)

	client := http.Client{}

	subRepo := subs.NewRepo(db)
	deliveryRepo := delivery.NewRepository(db)
	deliveryService := delivery.NewService(deliveryRepo, &client, statsClient, log)

	consumer := kafka.NewConsumer(cfg.KafkaCfg.Brokers, cfg.KafkaCfg.Topic)
	defer consumer.Close()

	wr := worker.NewWorker(consumer, log)
	log.Info("worker started")
	go func() {
		stream, err := statsClient.StreamStats(ctx, &statsv1.StreamStatsRequest{
			WebhookId:      "61604ceb-9cc1-4de1-8b98-992ac4d8c293", //TODO
			IntervalSecond: 2,
		})
		if err != nil {
			log.Error("stream", zap.Error(err))
			return
		}
		for {
			upd, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				log.Error("stream", zap.Error(err))
				return
			}
			if err != nil {
				log.Error("stream", zap.Error(err))
				return
			}
			fmt.Printf("total=%d failed=%d avg=%.1fms\n", upd.Total, upd.Failed, upd.AvgLatenct)
		}
	}()
	wr.Run(ctx, subRepo, deliveryService)
	return nil
}
