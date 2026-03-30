package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	segmentkafka "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/retry"
	"go.uber.org/zap"
)

type fakeConsumerClient struct {
	messages    []segmentkafka.Message
	commitErrs  []error
	commitCalls int
	mu          sync.Mutex
}

func (f *fakeConsumerClient) StartConsuming(ctx context.Context, out chan<- segmentkafka.Message, _ retry.Strategy) {
	go func() {
		defer close(out)
		for _, message := range f.messages {
			select {
			case <-ctx.Done():
				return
			case out <- message:
			}
		}
	}()
}

func (f *fakeConsumerClient) Commit(ctx context.Context, msg segmentkafka.Message) error {
	_ = ctx
	_ = msg

	f.mu.Lock()
	defer f.mu.Unlock()

	f.commitCalls++
	if len(f.commitErrs) == 0 {
		return nil
	}

	err := f.commitErrs[0]
	f.commitErrs = f.commitErrs[1:]
	return err
}

func (f *fakeConsumerClient) CommitCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.commitCalls
}

func TestConsumerRetriesHandlerAndCommitsAfterSuccess(t *testing.T) {
	client := &fakeConsumerClient{
		messages: []segmentkafka.Message{{Value: []byte(`{"image_id":"image-1"}`)}},
	}
	consumer := newConsumer(client, retry.Strategy{Attempts: 3, Backoff: 1}, zap.NewNop())

	attempts := 0
	err := consumer.Consume(context.Background(), func(ctx context.Context, imageID string) error {
		_ = ctx
		attempts++
		if imageID != "image-1" {
			t.Fatalf("unexpected image id: %s", imageID)
		}
		if attempts < 3 {
			return fmt.Errorf("temporary failure %d", attempts)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Consume returned error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 handler attempts, got %d", attempts)
	}

	if got := client.CommitCalls(); got != 1 {
		t.Fatalf("expected 1 commit, got %d", got)
	}
}

func TestConsumerDoesNotCommitWhenHandlerFails(t *testing.T) {
	client := &fakeConsumerClient{
		messages: []segmentkafka.Message{{Value: []byte(`{"image_id":"image-2"}`)}},
	}
	consumer := newConsumer(client, retry.Strategy{Attempts: 2, Backoff: 1}, zap.NewNop())

	attempts := 0
	err := consumer.Consume(context.Background(), func(ctx context.Context, imageID string) error {
		_ = ctx
		attempts++
		return errors.New("processing failed")
	})
	if err == nil {
		t.Fatal("expected Consume to return handler error")
	}

	if attempts != 2 {
		t.Fatalf("expected 2 handler attempts, got %d", attempts)
	}

	if got := client.CommitCalls(); got != 0 {
		t.Fatalf("expected 0 commits, got %d", got)
	}
}

func TestConsumerRetriesCommit(t *testing.T) {
	client := &fakeConsumerClient{
		messages:   []segmentkafka.Message{{Value: []byte(`{"image_id":"image-3"}`)}},
		commitErrs: []error{errors.New("commit failed")},
	}
	consumer := newConsumer(client, retry.Strategy{Attempts: 2, Backoff: 1}, zap.NewNop())

	err := consumer.Consume(context.Background(), func(ctx context.Context, imageID string) error {
		_ = ctx
		if imageID != "image-3" {
			t.Fatalf("unexpected image id: %s", imageID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Consume returned error: %v", err)
	}

	if got := client.CommitCalls(); got != 2 {
		t.Fatalf("expected 2 commit attempts, got %d", got)
	}
}

func TestConsumerCommitsMalformedMessages(t *testing.T) {
	client := &fakeConsumerClient{
		messages: []segmentkafka.Message{{Value: []byte(`not-json`)}},
	}
	consumer := newConsumer(client, retry.Strategy{Attempts: 1, Backoff: 1}, zap.NewNop())

	handlerCalled := false
	err := consumer.Consume(context.Background(), func(ctx context.Context, imageID string) error {
		_ = ctx
		_ = imageID
		handlerCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("Consume returned error: %v", err)
	}

	if handlerCalled {
		t.Fatal("expected handler to be skipped for malformed message")
	}

	if got := client.CommitCalls(); got != 1 {
		t.Fatalf("expected malformed message to be committed once, got %d", got)
	}
}
