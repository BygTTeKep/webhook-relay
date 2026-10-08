package relay

import (
	"context"
	"encoding/json"
	"time"
	"webhook-relay/internal/kafka"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Runner struct {
	repo     *Repository
	producer *kafka.Producer
	logger   *zap.Logger
}

func NewRunner(repo *Repository, producer *kafka.Producer, logger *zap.Logger) *Runner {
	return &Runner{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func idsOf(rows []Message) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, v := range rows {
		ids = append(ids, v.OutboxID)
	}
	return ids
}

func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.ticker(ctx); err != nil {
				r.logger.Error("relay tick failed", zap.Error(err))
			}
		}
	}
}

func (r *Runner) ticker(ctx context.Context) error {
	err := r.repo.WithTx(ctx, func(tx pgx.Tx) error {
		rows, err := r.repo.FetchPending(ctx, tx, 100)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		for _, row := range rows {
			body, err := json.Marshal(row)
			if err != nil {
				return err
			}
			if err := r.producer.Publish(ctx, "events", body); err != nil {
				return err
			}
		}
		ids := idsOf(rows)
		err = r.repo.MarkSend(ctx, tx, ids)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
