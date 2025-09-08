package pubsub

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type Pubsub interface {
	Publisher() Publisher
	Shutdown() error
}
type pubsubClient struct {
	client    *pubsub.Client
	publisher Publisher
}

func NewPubSubClient(ctx context.Context, projectID string, credentialsPath string, topicID string) (Pubsub, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("credentialsPath cannot be empty if specified for PubSub client")
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

	publisher, err := NewPublisher(ctx, client, topicID)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create publisher for topic '%s': %w", topicID, err)
	}

	return &pubsubClient{
		client:    client,
		publisher: publisher,
	}, nil
}

func (p *pubsubClient) Publisher() Publisher {
	return p.publisher
}

func (p *pubsubClient) Shutdown() error {
	if p.client != nil {
		return p.client.Close()
	}

	return nil
}
