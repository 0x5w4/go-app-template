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
)

var _ AuthService = (*authService)(nil)

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*entity.User, error)
	Refresh(ctx context.Context, req *RefreshRequest) (*entity.Token, error)
	Logout(ctx context.Context, req *LogoutRequest) error
	AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error)
}

type authService struct {
	config     *config.Config
	token      token.Token
	repository repository.Repository
	logger     logger.Logger
}

func NewAuthService(config *config.Config, token token.Token, repo repository.Repository, log logger.Logger) *authService {
	return &authService{
		config:     config,
		token:      token,
		repository: repo,
		logger:     log,
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
		_ = s.repository.Redis().DeleteUserAttempts(bgCtx, req.Username)
		_ = s.repository.Redis().DeleteIPAttempts(bgCtx, ip)
		_ = s.repository.Redis().DeleteBlockCount(bgCtx, ip)
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
			s.logger.Error().Msg("Access token has no JTI (ID), cannot blacklist")
		} else {
			err := s.repository.Redis().BlacklistToken(ctx, req.AccessTokenClaims.ID, atTTL)
			if err != nil {
				s.logger.Error().Msgf("Failed to blacklist access token: %v", err)
			}
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
			s.logger.Error().Msg("Refresh token has no JTI (ID), cannot blacklist")
		} else {
			err := s.repository.Redis().BlacklistToken(ctx, rtClaims.ID, rtTTL)
			if err != nil {
				s.logger.Error().Msgf("Failed to blacklist refresh token: %v", err)
			}
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
