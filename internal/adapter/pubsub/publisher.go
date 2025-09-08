package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
)

type Publisher interface {
	Publish(ctx context.Context, data []byte, attributes map[string]string) (string, error)
}

type publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

type CommandMessage struct {
	Command string          `json:"command"`
	Payload string          `json:"payload"`
	ID      uint            `json:"id"`
	Detail  string          `json:"detail"`
	Message *pubsub.Message `json:"-"`
}

func NewPublisher(ctx context.Context, client *pubsub.Client, topicID string) (Publisher, error) {
	topic := client.Topic(topicID)

	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topicID)
	}

	return &publisher{
		client: client,
		topic:  topic,
	}, nil
}

func (p *publisher) Publish(ctx context.Context, data []byte, attributes map[string]string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	return id, nil
}
