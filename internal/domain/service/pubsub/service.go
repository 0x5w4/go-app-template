package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/adapter/util/exception"
)

type PubsubService interface {
	SendToPublisher(ctx context.Context, image string, id uint, modelType string, filename, userLog string) error
}

type pubsubService struct {
	config *config.Config
	logger logger.Logger
	pubsub pubsub.Pubsub
}

func NewPubsubService(config *config.Config, logger logger.Logger, pubsub pubsub.Pubsub) PubsubService {
	return &pubsubService{
		config: config,
		logger: logger,
		pubsub: pubsub,
	}
}

func (s *pubsubService) SendToPublisher(ctx context.Context, image string, id uint, modelType string, filename, userLog string) error {
	url := s.config.HTTP.DomainName + "/api/v1/webhook/update-icon?id=" + fmt.Sprintf("%v", id) + "&type=" + fmt.Sprintf("%v", modelType)
	payload := PubImageReq{
		WebhookURL: url,
		Image:      image,
		Filename:   filename,
		FolderID:   s.config.Drive.IconFolderID,
	}

	payloadJSON, _ := json.Marshal(payload)
	msg := pubsub.CommandMessage{
		Command: "pub image",
		Payload: string(payloadJSON),
		Detail:  userLog,
	}
	msgJSON, _ := json.Marshal(msg)

	_, err := s.pubsub.Publisher().Publish(ctx, msgJSON, nil)
	if err != nil {
		return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to publish message")
	}

	return nil
}
