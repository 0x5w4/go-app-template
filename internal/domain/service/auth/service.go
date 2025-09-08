package auth

import (
	"context"

	"goapptemp/config"
	"goapptemp/internal/adapter/logger"
	repo "goapptemp/internal/adapter/repository"
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/adapter/util"
	"goapptemp/internal/adapter/util/exception"
	"goapptemp/internal/adapter/util/token"
	"goapptemp/internal/domain/entity"
	serror "goapptemp/internal/domain/service/error"
)

type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*entity.User, error)
	AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error)
}

type authService struct {
	config       *config.Config
	tokenManager token.Token
	repo         repo.Repository
	logger       logger.Logger
}

func NewAuthService(cfg *config.Config, tm token.Token, repo repo.Repository, logger logger.Logger) AuthService {
	return &authService{
		config:       cfg,
		tokenManager: tm,
		repo:         repo,
		logger:       logger,
	}
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*entity.User, error) {
	users, _, err := s.repo.MySQL().User().Find(ctx, &mysql.FilterUserPayload{Usernames: []string{req.User.Username}})
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	if len(users) == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeUserInvalidLogin, "Invalid username or password")
	}

	user := users[0]
	if err := util.CheckPassword(req.User.Password, user.Password); err != nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeUserInvalidLogin, "Invalid username or password")
	}

	token, err := s.tokenManager.Generate(user.ID)
	if err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to generate authentication token")
	}

	user.Token = &token

	return user, nil
}

func (s *authService) AuthorizationCheck(ctx context.Context, userID uint, permissionCode string) (bool, error) {
	if userID == 0 {
		return false, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "User id not provided")
	}

	user, err := s.repo.MySQL().User().FindByID(ctx, userID)
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
