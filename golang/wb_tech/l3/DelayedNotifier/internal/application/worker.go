package application

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/DelayedNotifier/internal/domain/interfaces"
	"go.uber.org/zap"
)

type NotifierWorker struct {
	service interfaces.NotifierService
	channel *rabbitmq.Channel
	queue   string
	logger  *zap.Logger
}

func NewNotifierWorker(service interfaces.NotifierService, channel *rabbitmq.Channel, queue string, logger *zap.Logger) *NotifierWorker {
	return &NotifierWorker{service: service, channel: channel, queue: queue, logger: logger}
}

func (w *NotifierWorker) Run() {
	msgs, err := w.channel.Consume(w.queue, "", false, false, false, false, nil)
	if err != nil {
		w.logger.Fatal("Failed to start worker", zap.Error(err))
	}
	w.logger.Info("Notifier worker started, waiting for messages")

	for delivery := range msgs {
		w.processMessage(delivery)
	}
	w.logger.Info("Notifier worker stopped")
}

func (w *NotifierWorker) processMessage(delivery amqp091.Delivery) {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(delivery.Body, &payload); err != nil {
		w.logger.Error("Failed to unmarshal message", zap.Error(err))
		err = delivery.Ack(false)
		if err != nil {
			w.logger.Error("Failed to ack message", zap.Error(err))
		}
		return
	}

	ctx := context.Background()
	note, err := w.service.GetNote(ctx, payload.ID)
	w.logger.Info("Worker processing note", zap.String("id", payload.ID), zap.String("status", note.Status))
	if err != nil {
		w.logger.Warn("Note not found in repo, skipping", zap.String("id", payload.ID), zap.Error(err))
		err = delivery.Ack(false)
		if err != nil {
			w.logger.Error("Failed to ack message", zap.Error(err))
		}
		return
	}

	if note.Status != "in_queue" {
		w.logger.Warn("Note not in queue yet, requeueing", zap.String("id", payload.ID))
		err = delivery.Nack(false, true)
		if err != nil {
			w.logger.Error("Failed to requeue message", zap.Error(err))
		}
		return
	}

	err = w.service.SendNotification(ctx, note)
	if err == nil {
		w.logger.Info("Notification sent successfully", zap.String("id", payload.ID))
		err = w.service.UpdateNoteStatus(ctx, payload.ID, "sent")
		if err != nil {
			w.logger.Error("Failed to update note status to sent", zap.String("id", payload.ID), zap.Error(err))
		}
		err = delivery.Ack(false)
		if err != nil {
			w.logger.Error("Failed to ack message", zap.Error(err))
		}
		return
	}

	w.logger.Error("Failed to send notification, will retry later...", zap.String("id", payload.ID), zap.Error(err))
	note.Retries++
	err = w.service.UpdateNoteRetries(ctx, payload.ID, note.Retries)
	if err != nil {
		w.logger.Error("Failed to update note retries", zap.String("id", payload.ID), zap.Error(err))
	}

	maxRetries := 5
	if note.Retries >= maxRetries {
		w.logger.Error("Max retries reached, marking as failed", zap.String("id", note.ID))
		err = w.service.UpdateNoteStatus(ctx, note.ID, "failed")
		if err != nil {
			w.logger.Error("Failed to update note status to failed", zap.String("id", note.ID), zap.Error(err))
		}
	} else {
		err = w.service.UpdateNoteStatus(ctx, note.ID, "pending")
		if err != nil {
			w.logger.Error("Failed to update note status to pending", zap.String("id", note.ID), zap.Error(err))
		}
	}
	_ = delivery.Ack(false)
}
