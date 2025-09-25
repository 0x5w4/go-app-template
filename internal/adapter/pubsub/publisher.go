package pubsub

import (
	"context"
	"goapptemp/pkg/logger"
	"time"

	pubsubClient "goapptemp/pkg/pubsub"

	"cloud.google.com/go/pubsub"
	cerrors "github.com/cockroachdb/errors"
)

type Publisher interface {
	Publish(ctx context.Context, data []byte, attributes map[string]string) (string, error)
}

type publisher struct {
	logger logger.Logger
	topic  *pubsub.Topic
}

func NewPublisher(logger logger.Logger, pubsub pubsubClient.Pubsub, topicID string) (*publisher, error) {
	topic, err := pubsub.NewPublisher(context.Background(), topicID)
	if err != nil {
		return nil, err
	}

	return &publisher{
		logger: logger,
		topic:  topic,
	}, nil
}

// TODO: Add log hook for success and failure.
func (p *publisher) Publish(ctx context.Context, data []byte, attributes map[string]string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})

	id, err := result.Get(ctx)
	if err != nil {
		return "", cerrors.Errorf("failed to publish message: %w", err)
	}

	return id, nil
}
