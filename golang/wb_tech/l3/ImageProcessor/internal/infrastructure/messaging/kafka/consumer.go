package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	segmentkafka "github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"go.uber.org/zap"
)

type consumerClient interface {
	StartConsuming(ctx context.Context, out chan<- segmentkafka.Message, strategy retry.Strategy)
	Commit(ctx context.Context, msg segmentkafka.Message) error
}

type Consumer struct {
	consumer consumerClient
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
	return newConsumer(wbfkafka.NewConsumer(brokers, topic, group), strategy, logger)
}

func newConsumer(consumer consumerClient, strategy retry.Strategy, logger *zap.Logger) *Consumer {
	return &Consumer{
		consumer: consumer,
		strategy: strategy,
		logger:   logger,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, string) error) error {
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

			shouldCommit, err := c.handleMessage(ctx, message, handler)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				return err
			}

			if !shouldCommit {
				continue
			}

			if err := c.commitWithRetry(ctx, message); err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				return err
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, message segmentkafka.Message, handler func(context.Context, string) error) (bool, error) {
	var job imageJobMessage
	if err := json.Unmarshal(message.Value, &job); err != nil {
		c.logger.Error("failed to decode kafka message", zap.Error(err))
		return true, nil
	}

	if err := c.handleWithRetry(ctx, job.ImageID, handler); err != nil {
		return false, fmt.Errorf("handle kafka message for image %q: %w", job.ImageID, err)
	}

	return true, nil
}

func (c *Consumer) handleWithRetry(ctx context.Context, imageID string, handler func(context.Context, string) error) error {
	attempts := normalizedAttempts(c.strategy.Attempts)
	delay := c.strategy.Delay
	backoff := normalizedBackoff(c.strategy.Backoff)

	for attempt := 1; attempt <= attempts; attempt++ {
		err := handler(ctx, imageID)
		if err == nil {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			return err
		}

		if attempt == attempts {
			return err
		}

		c.logger.Warn(
			"kafka job handler failed, retrying",
			zap.String("image_id", imageID),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", attempts),
			zap.Error(err),
		)

		if err := waitWithContext(ctx, delay); err != nil {
			return err
		}

		delay = nextDelay(delay, backoff)
	}

	return nil
}

func (c *Consumer) commitWithRetry(ctx context.Context, message segmentkafka.Message) error {
	attempts := normalizedAttempts(c.strategy.Attempts)
	delay := c.strategy.Delay
	backoff := normalizedBackoff(c.strategy.Backoff)

	for attempt := 1; attempt <= attempts; attempt++ {
		err := c.consumer.Commit(ctx, message)
		if err == nil {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			return err
		}

		if attempt == attempts {
			return fmt.Errorf("commit kafka message: %w", err)
		}

		c.logger.Warn(
			"failed to commit kafka message, retrying",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", attempts),
			zap.Error(err),
		)

		if err := waitWithContext(ctx, delay); err != nil {
			return err
		}

		delay = nextDelay(delay, backoff)
	}

	return nil
}

func normalizedAttempts(attempts int) int {
	if attempts <= 0 {
		return 1
	}
	return attempts
}

func normalizedBackoff(backoff float64) float64 {
	if backoff <= 0 {
		return 1
	}
	return backoff
}

func nextDelay(delay time.Duration, backoff float64) time.Duration {
	if delay <= 0 {
		return 0
	}
	return time.Duration(float64(delay) * backoff)
}

func waitWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
