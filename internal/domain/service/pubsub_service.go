package service

import (
	"context"
	"encoding/json"
	"goapptemp/config"
	"goapptemp/internal/adapter/pubsub"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"strconv"

	p "cloud.google.com/go/pubsub"
)

var _ PubsubService = (*pubsubService)(nil)

type PubsubService interface {
	SendToPublisher(ctx context.Context, image string, id uint, modelType string, filename, userLog string) error
}

type pubsubService struct {
	config    *config.Config
	logger    logger.Logger
	publisher pubsub.Publisher
}

func NewPubsubService(config *config.Config, logger logger.Logger, publisher pubsub.Publisher) *pubsubService {
	return &pubsubService{
		config:    config,
		logger:    logger,
		publisher: publisher,
	}
}

type CommandMessage struct {
	Command string     `json:"command"`
	Payload string     `json:"payload"`
	ID      uint       `json:"id"`
	Detail  string     `json:"detail"`
	Message *p.Message `json:"-"`
}

type PubImageReq struct {
	WebhookURL string `json:"url_webhook"`
	Image      string `json:"image"`
	FolderID   string `json:"folder_id"`
	Filename   string `json:"filename"`
}

func (s *pubsubService) SendToPublisher(ctx context.Context, image string, id uint, modelType string, filename, userLog string) error {
	url := s.config.HTTP.DomainName + "/api/v1/webhook/update-icon?id=" + strconv.FormatUint(uint64(id), 10) + "&type=" + modelType
	payload := PubImageReq{
		WebhookURL: url,
		Image:      image,
		Filename:   filename,
		FolderID:   s.config.Drive.IconFolderID,
	}

	payloadJSON, _ := json.Marshal(payload)
	msg := CommandMessage{
		Command: "pub image",
		Payload: string(payloadJSON),
		Detail:  userLog,
	}
	msgJSON, _ := json.Marshal(msg)

	_, err := s.publisher.Publish(ctx, msgJSON, nil)
	if err != nil {
		return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to publish message")
	}

	return nil
}
