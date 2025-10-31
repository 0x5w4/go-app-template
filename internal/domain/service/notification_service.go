package service

import (
	"context"
	"fmt"
	"goapptemp/pkg/logger"
)

type NotificationService interface {
	SendPasswordResetEmail(ctx context.Context, userEmail, resetLink string) error
	SendPasswordResetSuccessEmail(ctx context.Context, userEmail string) error
}

type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

var _ NotificationService = (*notificationService)(nil)

type notificationService struct {
	emailSender EmailSender
	logger      logger.Logger
}

func NewNotificationService(emailSender EmailSender, log logger.Logger) NotificationService {
	return &notificationService{
		emailSender: emailSender,
		logger:      log,
	}
}

func (s *notificationService) SendPasswordResetEmail(ctx context.Context, userEmail, resetLink string) error {
	subject := "Reset Your Password"

	body := fmt.Sprintf(`
		<p>Hello,</p>
		<p>You requested a password reset. Click the link below to set a new password:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link is valid for 15 minutes.</p>
		<p>If you did not request this, please ignore this email.</p>
	`, resetLink)

	err := s.emailSender.SendEmail(ctx, userEmail, subject, body)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Failed to send password reset email to %s", userEmail)
		return err
	}

	s.logger.Info().Msgf("Password reset email sent to %s", userEmail)

	return nil
}

func (s *notificationService) SendPasswordResetSuccessEmail(ctx context.Context, userEmail string) error {
	subject := "Your Password Has Been Changed"

	body := `
		<p>Hello,</p>
		<p>This is a confirmation that the password for your account has just been changed.</p>
		<p>If you did not make this change, please contact our support team immediately.</p>
	`

	err := s.emailSender.SendEmail(ctx, userEmail, subject, body)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Failed to send password reset success email to %s", userEmail)
		return err
	}

	s.logger.Info().Msgf("Password reset success email sent to %s", userEmail)

	return nil
}
