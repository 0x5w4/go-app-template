package service

import (
	"context"
	"goapptemp/config"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	repo "goapptemp/internal/adapter/repository"

	serror "goapptemp/internal/domain/service/error"
)

var _ RoleService = (*roleService)(nil)

type RoleService interface {
	Create(ctx context.Context, req *CreateRoleRequest) (*entity.Role, error)
	Update(ctx context.Context, req *UpdateRoleRequest) (*entity.Role, error)
	Delete(ctx context.Context, req *DeleteRoleRequest) error
	Find(ctx context.Context, req *FindRolesRequest) ([]*entity.Role, int, error)
	FindOne(ctx context.Context, req *FindOneRoleRequest) (*entity.Role, error)
}

type roleService struct {
	config *config.Config
	repo   repo.Repository
	logger logger.Logger
	auth   AuthService
}

func NewRoleService(config *config.Config, repo repo.Repository, logger logger.Logger, auth AuthService) *roleService {
	return &roleService{
		config: config,
		repo:   repo,
		logger: logger,
		auth:   auth,
	}
}

type CreateRoleRequest struct {
	AuthParams *AuthParams
	Role       *entity.Role
}

func (s *roleService) Create(ctx context.Context, req *CreateRoleRequest) (*entity.Role, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "ROLE.CREATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.Role == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Role data cannot be nil")
	}

	var role *entity.Role

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		var err error

		role, err = txRepo.Role().Create(ctx, req.Role)
		if err != nil {
			return err
		}

		if len(req.Role.PermissionIDs) != 0 {
			_, err = txRepo.Role().AttachPermissions(ctx, role.ID, req.Role.PermissionIDs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	role, err = s.repo.MySQL().Role().FindByID(ctx, role.ID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return role, nil
}

type UpdateRoleRequest struct {
	AuthParams *AuthParams
	Update     *mysqlrepository.UpdateRolePayload
}

func (s *roleService) Update(ctx context.Context, req *UpdateRoleRequest) (*entity.Role, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "ROLE.UPDATE")
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
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Role ID required for update")
	}

	var role *entity.Role

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		var err error

		role, err = txRepo.Role().Update(ctx, req.Update)
		if err != nil {
			return err
		}

		if req.Update.PermissionIDs != nil {
			IDsMap := make(map[uint]bool)

			for _, IDPtr := range req.Update.PermissionIDs {
				if IDPtr != nil {
					IDsMap[*IDPtr] = true
				}
			}

			permissionIDs := make([]uint, 0, len(IDsMap))
			for IDMap := range IDsMap {
				permissionIDs = append(permissionIDs, IDMap)
			}

			_, err = txRepo.Role().SyncPermissions(ctx, role.ID, permissionIDs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	role, err = s.repo.MySQL().Role().FindByID(ctx, role.ID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return role, nil
}

type DeleteRoleRequest struct {
	AuthParams *AuthParams
	RoleID     uint
}

func (s *roleService) Delete(ctx context.Context, req *DeleteRoleRequest) error {
	if req.AuthParams.AccessTokenClaims == nil {
		return exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "ROLE.DELETE")
	if err != nil {
		return err
	}

	if !ok {
		return exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.RoleID == 0 {
		return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Role ID cannot be zero")
	}

	err = s.repo.MySQL().Role().Delete(ctx, req.RoleID)
	if err != nil {
		return serror.TranslateRepoError(err)
	}

	return nil
}

type FindRolesRequest struct {
	AuthParams *AuthParams
	Filter     *mysqlrepository.FilterRolePayload
}

func (s *roleService) Find(ctx context.Context, req *FindRolesRequest) ([]*entity.Role, int, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, 0, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "ROLE.READ")
	if err != nil {
		return nil, 0, err
	}

	if !ok {
		return nil, 0, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	roles, totalCount, err := s.repo.MySQL().Role().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return roles, totalCount, nil
}

type FindOneRoleRequest struct {
	AuthParams *AuthParams
	RoleID     uint
}

func (s *roleService) FindOne(ctx context.Context, req *FindOneRoleRequest) (*entity.Role, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "ROLE.READ")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.RoleID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Role ID required for find one")
	}

	role, err := s.repo.MySQL().Role().FindByID(ctx, req.RoleID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return role, nil
}
