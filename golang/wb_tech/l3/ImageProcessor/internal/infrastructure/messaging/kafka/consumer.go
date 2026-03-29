package kafka

import (
	"context"
	"encoding/json"
	"errors"

	segmentkafka "github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
	"go.uber.org/zap"
)

type Consumer struct {
	consumer *wbfkafka.Consumer
	strategy retry.Strategy
	logger   *zap.Logger
}

func NewConsumer(
	brokers []string,
	topic string,
	group string,
	strategy retry.Strategy,
	logger *zap.Logger,
) *Consumer {
	return &Consumer{
		consumer: wbfkafka.NewConsumer(brokers, topic, group),
		strategy: strategy,
		logger:   logger,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, dto.ImageJob) error) error {
	messageCh := make(chan segmentkafka.Message)

	c.consumer.StartConsuming(ctx, messageCh, c.strategy)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-messageCh:
			if !ok {
				return nil
			}

			var job dto.ImageJob
			if err := json.Unmarshal(message.Value, &job); err != nil {
				c.logger.Error("failed to decode kafka message", zap.Error(err))
			} else if err := handler(ctx, job); err != nil && !errors.Is(err, context.Canceled) {
				c.logger.Error("kafka job handler failed", zap.String("image_id", job.ImageID), zap.Error(err))
			}

			if err := c.consumer.Commit(ctx, message); err != nil && !errors.Is(err, context.Canceled) {
				c.logger.Error("failed to commit kafka message", zap.Error(err))
			}
		}
	}
}
