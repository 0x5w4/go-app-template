package pubsub

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type Pubsub interface {
	NewPublisher(ctx context.Context, topicID string) (*pubsub.Topic, error)
	Shutdown() error
}
type pubsubClient struct {
	client *pubsub.Client
}

func NewPubsub(ctx context.Context, projectID string, credentialsPath string) (Pubsub, error) {
	if credentialsPath == "" {
		return nil, errors.New("credentialsPath cannot be empty if specified for PubSub client")
	}

	fileInfo, err := os.Stat(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials file not found at path '%s': %w", credentialsPath, err)
		}

		return nil, fmt.Errorf("error accessing credentials file at path '%s': %w", credentialsPath, err)
	}

	if fileInfo.IsDir() {
		return nil, fmt.Errorf("credentialsPath '%s' points to a directory, not a file", credentialsPath)
	}

	clientOpts := []option.ClientOption{option.WithCredentialsFile(credentialsPath)}

	client, err := pubsub.NewClient(ctx, projectID, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client with credentials file '%s': %w", credentialsPath, err)
	}

	return &pubsubClient{
		client: client,
	}, nil
}

func (p *pubsubClient) NewPublisher(ctx context.Context, topicID string) (*pubsub.Topic, error) {
	topic := p.client.Topic(topicID)

	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("topic %q does not exist", topicID)
	}

	return topic, nil
}

func (p *pubsubClient) Shutdown() error {
	if p.client != nil {
		return p.client.Close()
	}

	return nil
}
