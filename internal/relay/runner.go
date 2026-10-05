package relay

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
	"webhook-relay/internal/kafka"
)

type Runner struct {
	repo *Repository
	producer *kafka.Producer
}

func NewRunner(repo *Repository, producer *kafka.Producer) *Runner {
	return &Runner{
		repo: repo,
		producer: producer,
	}
}

func idsOf(rows []OutBoxEvents) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, v := range rows {
		ids = append(ids, v.ID)
	}
	return ids
}

func (r *Runner) Run(ctx context.Context) {
	ticker :=time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.ticker(ctx); err != nil {
				slog.Error("relay tick failed", "err", err)
			}
		}
	}
}

func (r *Runner) ticker(ctx context.Context) error {
	r.repo.WithTx(ctx, func(tx *sql.Tx) error {
		rows, err := r.repo.FetchPending(ctx, tx, 100)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if err := r.producer.Publish(ctx, "events", row.Payload); err != nil {
				return err
			}
		}
		ids := idsOf(rows)
		err = r.repo.MarkSend(ctx, tx, ids)
		if err != nil {
			return  err
		}
		return  nil
	})
	return  nil;
}