package service

import (
	"context"
	"fmt"
	"goapptemp/config"
	"goapptemp/constant"
	"goapptemp/internal/adapter/repository"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	serror "goapptemp/internal/domain/service/error"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/internal/shared/token"
	"goapptemp/pkg/logger"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
)

var _ AuthService = (*authService)(nil)

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*entity.User, error)
	Refresh(ctx context.Context, req *RefreshRequest) (*entity.Token, error)
	Logout(ctx context.Context, req *LogoutRequest) error
	AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error)
	ForgetPassword(ctx context.Context, req *ForgetPasswordRequest) error
	VerifyResetToken(ctx context.Context, req *VerifyResetTokenRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
}

type authService struct {
	config          *config.Config
	token           token.Token
	repository      repository.Repository
	logger          logger.Logger
	notificationSvc NotificationService
}

func NewAuthService(
	config *config.Config,
	token token.Token,
	repo repository.Repository,
	log logger.Logger,
	notificationSvc NotificationService,
) *authService {
	return &authService{
		config:          config,
		token:           token,
		repository:      repo,
		logger:          log,
		notificationSvc: notificationSvc,
	}
}

type AuthParams struct {
	AccessToken       string
	AccessTokenClaims *token.AccessTokenClaims
}

type LoginRequest struct {
	Username string
	Password string
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*entity.User, error) {
	var (
		user                  *entity.User
		passwordHashToCompare string
		loginSuccessful       = false
	)

	ip, _ := ctx.Value(constant.CtxKeyRequestIP).(string)

	errGenericLogin := exception.New(exception.TypeBadRequest, exception.CodeUserInvalidLogin, "Invalid username or password")

	isLocked, err := s.repository.Redis().CheckLockedUserExists(ctx, req.Username)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	users, _, err := s.repository.MySQL().User().Find(ctx, &mysqlrepository.FilterUserPayload{Usernames: []string{req.Username}})
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	if len(users) == 0 || users[0] == nil || isLocked {
		passwordHashToCompare = constant.DummyPasswordHash
		user = nil
	} else {
		user = users[0]
		passwordHashToCompare = user.Password
	}

	errPass := shared.CheckPassword(req.Password, passwordHashToCompare)
	if errPass == nil {
		if user != nil && !isLocked {
			loginSuccessful = true
		}
	}

	if !loginSuccessful {
		go func() {
			bgCtx := context.Background()

			errUser := s.repository.Redis().RecordUserFailure(bgCtx, req.Username)
			if errUser != nil {
				s.logger.Error().Msgf("Failed to record user failure: %v", errUser)
			}

			_, _, errIP := s.repository.Redis().RecordIPFailure(bgCtx, ip)
			if errIP != nil {
				s.logger.Error().Msgf(fmt.Sprintf("Failed to record IP failure: %v", errIP))
			}
		}()

		return nil, errGenericLogin
	}

	go func() {
		bgCtx := context.Background()
		if err := s.repository.Redis().DeleteUserAttempts(bgCtx, req.Username); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete user attempts for %s", req.Username)
		}

		if err := s.repository.Redis().DeleteIPAttempts(bgCtx, ip); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete IP attempts for %s", ip)
		}

		if err := s.repository.Redis().DeleteBlockCount(bgCtx, ip); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete block count for %s", ip)
		}
	}()

	accessToken, accessExpiresAt, err := s.token.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "failed to generate access token")
	}

	refreshToken, refreshExpiresAt, err := s.token.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "failed to generate refresh token")
	}

	user.Token = &entity.Token{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
		TokenType:             constant.TokenType,
	}

	return user, nil
}

type RefreshRequest struct {
	RefreshToken string
}

func (s *authService) Refresh(ctx context.Context, req *RefreshRequest) (*entity.Token, error) {
	refreshTokenClaims, err := s.token.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeUnauthorized, exception.CodeUnauthorized, "invalid refresh token")
	}

	user, err := s.repository.MySQL().User().FindByID(ctx, refreshTokenClaims.UserID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	accessToken, accessExpiresAt, err := s.token.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "failed to generate access token")
	}

	refreshToken, refreshExpiresAt, err := s.token.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "failed to generate refresh token")
	}

	return &entity.Token{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
		TokenType:             constant.TokenType,
	}, nil
}

type LogoutRequest struct {
	AccessTokenClaims *token.AccessTokenClaims
	RefreshToken      string
}

func (s *authService) Logout(ctx context.Context, req *LogoutRequest) error {
	atExpiresAt := req.AccessTokenClaims.ExpiresAt.Time
	atTTL := time.Until(atExpiresAt)

	if atTTL > 0 {
		if req.AccessTokenClaims.ID == "" {
			return exception.New(exception.TypeInternalError, "JTI_MISSING", "Access token has no JTI (ID), cannot blacklist")
		}

		err := s.repository.Redis().BlacklistToken(ctx, req.AccessTokenClaims.ID, atTTL)
		if err != nil {
			return serror.TranslateRepoError(err)
		}
	}

	rtClaims, err := s.token.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Warn().Msgf("Invalid refresh token provided during logout: %v", err)
		return nil
	}

	if rtClaims.UserID != req.AccessTokenClaims.UserID {
		s.logger.Warn().Msgf("Logout attempt with mismatched tokens. UserID %d vs %d",
			req.AccessTokenClaims.UserID, rtClaims.UserID)

		return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "token mismatch")
	}

	rtExpiresAt := rtClaims.ExpiresAt.Time
	rtTTL := time.Until(rtExpiresAt)

	if rtTTL > 0 {
		if rtClaims.ID == "" {
			return exception.New(exception.TypeInternalError, "JTI_MISSING", "Refresh token has no JTI (ID), cannot blacklist")
		}

		err := s.repository.Redis().BlacklistToken(ctx, rtClaims.ID, rtTTL)
		if err != nil {
			return serror.TranslateRepoError(err)
		}
	}

	return nil
}

func (s *authService) AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error) {
	if userID == 0 {
		return false, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "User id not provided")
	}

	hasPermission, err := s.repository.MySQL().User().HasPermission(ctx, userID, permissionCode)
	if err != nil {
		return false, serror.TranslateRepoError(err)
	}

	return hasPermission, nil
}

type ForgetPasswordRequest struct {
	Email string
}

func (s *authService) ForgetPassword(ctx context.Context, req *ForgetPasswordRequest) error {
	users, _, err := s.repository.MySQL().User().Find(ctx, &mysqlrepository.FilterUserPayload{
		Emails: []string{req.Email},
	})
	if err != nil {
		s.logger.Error().Err(err).Msgf("Failed to find user by email for password reset")

		return nil
	}

	if len(users) == 0 || users[0] == nil {
		s.logger.Warn().Msgf("Password reset attempt for non-existent email: %s", req.Email)
		return nil
	}

	user := users[0]
	resetToken := uuid.NewString()
	ttl := time.Minute * 15

	err = s.repository.Redis().StoreResetToken(ctx, resetToken, user.ID, ttl)
	if err != nil {
		return serror.TranslateRepoError(err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s",
		s.config.App.FrontendURL,
		resetToken,
	)

	go func() {
		bgCtx := context.Background()

		err := s.notificationSvc.SendPasswordResetEmail(bgCtx, user.Email, resetLink)
		if err != nil {
			s.logger.Error().Err(err).Msgf("Failed to send password reset email to user: %d", user.ID)
		}
	}()

	return nil
}

type VerifyResetTokenRequest struct {
	Token string
}

func (s *authService) VerifyResetToken(ctx context.Context, req *VerifyResetTokenRequest) error {
	_, err := s.repository.Redis().GetUserIDFromResetToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			return exception.New(exception.TypeBadRequest, "INVALID_TOKEN", "Invalid or expired reset token")
		}

		return serror.TranslateRepoError(err)
	}

	return nil
}

type ResetPasswordRequest struct {
	Token       string
	NewPassword string
}

func (s *authService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	userID, err := s.repository.Redis().GetUserIDFromResetToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, exception.ErrNotFound) {
			return exception.New(exception.TypeBadRequest, "INVALID_TOKEN", "Invalid or expired reset token")
		}

		return serror.TranslateRepoError(err)
	}

	if len(req.NewPassword) < constant.MinPasswordLength {
		return exception.Newf(exception.TypeBadRequest, exception.CodeValidationFailed, "Password must be at least %d characters long", constant.MinPasswordLength)
	}

	user, err := s.repository.MySQL().User().FindByID(ctx, userID)
	if err != nil {
		return serror.TranslateRepoError(err)
	}

	hashedPassword, err := shared.HashPassword(req.NewPassword)
	if err != nil {
		return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "failed to hash new password")
	}

	updatePayload := &mysqlrepository.UpdateUserPayload{
		ID:       userID,
		Password: &hashedPassword,
	}
	if _, err := s.repository.MySQL().User().Update(ctx, updatePayload); err != nil {
		return serror.TranslateRepoError(err)
	}

	go func() {
		bgCtx := context.Background()

		err := s.repository.Redis().DeleteResetToken(bgCtx, req.Token)
		if err != nil {
			s.logger.Error().Err(err).Msgf("CRITICAL: Failed to delete reset token for user %d. Token: %s", userID, req.Token)
		}
	}()

	go func() {
		bgCtx := context.Background()
		ip, _ := ctx.Value(constant.CtxKeyRequestIP).(string)

		if err := s.repository.Redis().DeleteUserAttempts(bgCtx, user.Username); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete user attempts for %s", user.Username)
		}

		if err := s.repository.Redis().DeleteIPAttempts(bgCtx, ip); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete IP attempts for %s", ip)
		}

		if err := s.repository.Redis().DeleteBlockCount(bgCtx, ip); err != nil {
			s.logger.Error().Err(err).Msgf("Failed to delete block count for %s", ip)
		}
	}()

	go func() {
		bgCtx := context.Background()

		err := s.notificationSvc.SendPasswordResetSuccessEmail(bgCtx, user.Email)
		if err != nil {
			s.logger.Error().Err(err).Msgf("Failed to send password reset success email to user: %d", user.ID)
		}
	}()

	return nil
}
