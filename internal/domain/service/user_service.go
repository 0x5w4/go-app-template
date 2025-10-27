package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/internal/adapter/repository"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	serror "goapptemp/internal/domain/service/error"
)

var _ UserService = (*userService)(nil)

type UserService interface {
	Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error)
	Update(ctx context.Context, req *UpdateUserRequest) (*entity.User, error)
	Delete(ctx context.Context, req *DeleteUserRequest) error
	Find(ctx context.Context, req *FindUserRequest) ([]*entity.User, int, error)
	FindOne(ctx context.Context, req *FindOneUserRequest) (*entity.User, error)
}

type userService struct {
	config *config.Config
	repo   repository.Repository
	logger logger.Logger
	auth   AuthService
}

func NewUserService(config *config.Config, repo repository.Repository, logger logger.Logger, auth AuthService) *userService {
	return &userService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

type CreateUserRequest struct {
	AuthParams *AuthParams
	User       *entity.User
}

func (s *userService) Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "USER.CREATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.User == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "User data cannot be nil")
	}

	if err := req.User.SetPassword(req.User.Password); err != nil {
		return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to hash user password during create")
	}

	var user *entity.User

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		var err error

		user, err = txRepo.User().Create(ctx, req.User)
		if err != nil {
			return err
		}

		if len(req.User.RoleIDs) != 0 {
			_, err = txRepo.User().AttachRoles(ctx, user.ID, req.User.RoleIDs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	user, err = s.repo.MySQL().User().FindByID(ctx, user.ID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return user, nil
}

type UpdateUserRequest struct {
	AuthParams *AuthParams
	Update     *mysqlrepository.UpdateUserPayload
}

func (s *userService) Update(ctx context.Context, req *UpdateUserRequest) (*entity.User, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "USER.UPDATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.Update == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Update payload cannot be nil")
	}

	if req.Update.ID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "User ID required for update")
	}

	if req.Update.Password != nil && *req.Update.Password != "" {
		hashedPassword, err := shared.HashPassword(*req.Update.Password)
		if err != nil {
			return nil, exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to hash user password during update")
		}

		req.Update.Password = &hashedPassword
	} else {
		req.Update.Password = nil
	}

	var user *entity.User

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		var err error

		user, err = txRepo.User().Update(ctx, req.Update)
		if err != nil {
			return err
		}

		if req.Update.RoleIDs != nil {
			IDsMap := make(map[uint]bool)

			for _, IDPtr := range req.Update.RoleIDs {
				if IDPtr != nil {
					IDsMap[*IDPtr] = true
				}
			}

			roleIDs := make([]uint, 0, len(IDsMap))
			for IDMap := range IDsMap {
				roleIDs = append(roleIDs, IDMap)
			}

			_, err = txRepo.User().SyncRoles(ctx, user.ID, roleIDs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	user, err = s.repo.MySQL().User().FindByID(ctx, user.ID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return user, nil
}

type DeleteUserRequest struct {
	AuthParams *AuthParams
	UserID     uint
}

func (s *userService) Delete(ctx context.Context, req *DeleteUserRequest) error {
	if req.AuthParams.AccessTokenClaims == nil {
		return exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "USER.DELETE")
	if err != nil {
		return err
	}

	if !ok {
		return exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.UserID == 0 {
		return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "User ID cannot be zero")
	}

	if req.AuthParams.AccessTokenClaims.UserID == req.UserID {
		return exception.New(exception.TypeForbidden, exception.CodeForbidden, "User cannot delete their own account")
	}

	if err := s.repo.MySQL().User().Delete(ctx, req.UserID); err != nil {
		return serror.TranslateRepoError(err)
	}

	return nil
}

type FindUserRequest struct {
	AuthParams *AuthParams
	UserFilter *mysqlrepository.FilterUserPayload
}

func (s *userService) Find(ctx context.Context, req *FindUserRequest) ([]*entity.User, int, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, 0, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "USER.READ")
	if err != nil {
		return nil, 0, err
	}

	if !ok {
		return nil, 0, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	users, totalCount, err := s.repo.MySQL().User().Find(ctx, req.UserFilter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	for _, user := range users {
		user.Password = ""
	}

	return users, totalCount, nil
}

type FindOneUserRequest struct {
	AuthParams *AuthParams
	UserID     uint
}

func (s *userService) FindOne(ctx context.Context, req *FindOneUserRequest) (*entity.User, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "USER.READ")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.UserID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "User ID required for find one")
	}

	user, err := s.repo.MySQL().User().FindByID(ctx, req.UserID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	user.Password = ""

	return user, nil
}
