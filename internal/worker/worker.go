// Читаем сообщения из Outbox
// Отправляем сообщения всем подписчмкам
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"webhook-relay/internal/delivery"
	"webhook-relay/internal/events"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/relay"
	"webhook-relay/internal/subs"
)

type Worker struct {
	*kafka.Consumer
}

func NewWorker(c *kafka.Consumer) *Worker {
	return &Worker{
		c,
	}
}

func (w *Worker) Run(ctx context.Context, subsRepo subs.RepoInterface, d *delivery.Service) {
	for {
		msg, err := w.Reader.FetchMessage(ctx)
		slog.Info("message: ", msg)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Error("fetch message failed", "err", err)
			continue
		}
		var e relay.Message
		err = json.Unmarshal(msg.Value, &e)
		if err != nil {
			slog.Error("unmarshal event failed", "err", err)
			continue 
		}
		subs, err := subsRepo.FindByEventType(ctx, events.EventType(e.EventType))
		if err != nil {
			slog.Error("find subscriptions failed", "err", err)
			continue
		}
		for _, sub := range subs {
			// delivery 
			if err := d.Delivery(ctx, e, sub); err != nil {
				slog.Error("delivery failed", "sub", sub.ID, "err", err)
			}
		}
		if err := w.Reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit failed", "err", err)
		}
	}
}