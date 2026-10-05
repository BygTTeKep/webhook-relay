package relay

import (
	"context"
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
		cfg, err := config.LoadConfig("")
	if err != nil {
		stop()
	}
	db, err := database.NewPG(&cfg.DBCfg)
	if err != nil {
		stop()
	}
	relayRepo := relay.NewRepository(db)
	producer := kafka.NewProducer([]string{}, "events")
	runner := relay.NewRunner(relayRepo, producer)
	runner.Run(ctx)
	defer producer.Close()
}