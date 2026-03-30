package kafka

import (
	"context"
	"encoding/json"

	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type imageJobMessage struct {
	ImageID string `json:"image_id"`
}

type Publisher struct {
	producer *wbfkafka.Producer
	strategy retry.Strategy
}

func NewPublisher(brokers []string, topic string, strategy retry.Strategy) *Publisher {
	return &Publisher{
		producer: wbfkafka.NewProducer(brokers, topic),
		strategy: strategy,
	}
}

func (p *Publisher) Publish(ctx context.Context, imageID string) error {
	payload, err := json.Marshal(imageJobMessage{ImageID: imageID})
	if err != nil {
		return err
	}

	return p.producer.SendWithRetry(ctx, p.strategy, []byte(imageID), payload)
}
