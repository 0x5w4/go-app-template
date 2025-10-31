package gmailsender

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var _ GmailSender = (*gmailSender)(nil)

type GmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type gmailSender struct {
	gmailService *gmail.Service
	senderEmail  string
}

func NewGmailSender(ctx context.Context, credPath, tokenPath, senderEmail string) (*gmailSender, error) {
	b, err := os.ReadFile(credPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse client secret file to config")
	}

	client, err := getClient(config, tokenPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve Gmail client")
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve Gmail service")
	}

	return &gmailSender{
		gmailService: srv,
		senderEmail:  senderEmail,
	}, nil
}

func getClient(config *oauth2.Config, tokenFile string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		return nil, errors.Newf("cannot get token from file. Please run auth flow. %v", err)
	}

	return config.Client(context.Background(), tok), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

func (s *gmailSender) SendEmail(ctx context.Context, to, subject, body string) error {
	messageStr := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-version: 1.0;\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"+
			"%s",
		s.senderEmail, to, subject, body,
	)

	messageBytes := []byte(messageStr)
	encodedMessage := base64.URLEncoding.EncodeToString(messageBytes)

	gMessage := &gmail.Message{
		Raw: encodedMessage,
	}

	_, err := s.gmailService.Users.Messages.Send("me", gMessage).Context(ctx).Do()
	if err != nil {
		return errors.Wrap(err, "failed to send email via Gmail API")
	}

	return nil
}
