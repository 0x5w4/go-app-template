package pubsub

type PubImageReq struct {
	WebhookURL string `json:"url_webhook"`
	Image      string `json:"image"`
	FolderID   string `json:"folder_id"`
	Filename   string `json:"filename"`
}
