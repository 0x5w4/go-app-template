package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/constant"
	"goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/internal/shared/token"
	"goapptemp/pkg/logger"

	serror "goapptemp/internal/domain/service/error"
)

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*entity.User, error)
	AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error)
}

type authService struct {
	config     *config.Config
	token      token.Token
	repository repository.Repository
	logger     logger.Logger
}

func NewAuthService(config *config.Config, token token.Token, repo repository.Repository, log logger.Logger) AuthService {
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

type RefreshRequest struct {
	RefreshToken string
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*entity.User, error) {
	users, _, err := s.repository.MySQL().User().Find(ctx, &mysql.FilterUserPayload{Usernames: []string{req.Username}})
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	if len(users) == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeUserInvalidLogin, "Invalid username or password")
	}

	user := users[0]
	if err := shared.CheckPassword(req.Password, user.Password); err != nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeUserInvalidLogin, "Invalid username or password")
	}

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
		TokenType:             constant.TOKEN_TYPE,
	}

	return user, nil
}

func (s *authService) AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error) {
	if userID == 0 {
		return false, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "User id not provided")
	}

	user, err := s.repository.MySQL().User().FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, exception.New(exception.TypeNotFound, exception.CodeNotFound, "User not found")
	}

	for _, role := range user.Roles {
		if role.SuperAdmin {
			return true, nil
		}

		for _, permission := range role.Permissions {
			if permission.Code == permissionCode {
				return true, nil
			}
		}
	}

	return false, nil
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
		TokenType:             constant.TOKEN_TYPE,
	}, nil
}
