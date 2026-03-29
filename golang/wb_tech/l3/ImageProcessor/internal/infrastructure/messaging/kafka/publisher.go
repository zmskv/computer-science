package kafka

import (
	"context"
	"encoding/json"

	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/zmskv/computer-science/golang/wb_tech/l3/ImageProcessor/internal/application/dto"
)

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

func (p *Publisher) Publish(ctx context.Context, job dto.ImageJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.producer.SendWithRetry(ctx, p.strategy, []byte(job.ImageID), payload)
}
