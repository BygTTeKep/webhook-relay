// Читаем сообщения из Outbox
// Отправляем сообщения всем подписчмкам
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"webhook-relay/internal/delivery"
	"webhook-relay/internal/events"
	"webhook-relay/internal/kafka"
	"webhook-relay/internal/relay"
	"webhook-relay/internal/subs"

	"go.uber.org/zap"
)

type Worker struct {
	c      *kafka.Consumer
	logger *zap.Logger
}

func NewWorker(c *kafka.Consumer, logger *zap.Logger) *Worker {
	return &Worker{
		c:      c,
		logger: logger,
	}
}

func (w *Worker) Run(ctx context.Context, subsRepo subs.RepoInterface, d *delivery.Service) {
	for {
		msg, err := w.c.Reader.FetchMessage(ctx)
		w.logger.Info("message: ", zap.Any("mesage", msg))
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			w.logger.Error("fetch message failed", zap.Error(err))
			continue
		}
		var e relay.Message
		err = json.Unmarshal(msg.Value, &e)
		if err != nil {
			w.logger.Error("unmarshal event failed", zap.Error(err))
			continue
		}
		subs, err := subsRepo.FindByEventType(ctx, events.EventType(e.EventType))
		if err != nil {
			w.logger.Error("find subscriptions failed", zap.Error(err))
			continue
		}
		for _, sub := range subs {
			// delivery
			if err := d.Delivery(ctx, e, sub); err != nil {
				w.logger.Error("delivery failed", []zap.Field{zap.String("sub", sub.ID), zap.Error(err)}...)
			}
		}
		if err := w.c.Reader.CommitMessages(ctx, msg); err != nil {
			w.logger.Error("commit failed", zap.Error(err))
		}
	}
}
