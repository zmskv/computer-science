package kafka

import (
	"context"
	"errors"

	segmentkafka "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/retry"
)

func EnsureTopic(
	ctx context.Context,
	brokers []string,
	topic string,
	partitions int,
	replicationFactor int,
	strategy retry.Strategy,
) error {
	if len(brokers) == 0 {
		return errors.New("at least one kafka broker is required")
	}

	if topic == "" {
		return errors.New("kafka topic is required")
	}

	return retry.Do(func() error {
		conn, err := segmentkafka.DialLeader(ctx, "tcp", brokers[0], topic, 0)
		if err != nil {
			return err
		}
		return conn.Close()
	}, strategy)
}
