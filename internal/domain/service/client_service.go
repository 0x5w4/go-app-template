package service

import (
	"context"
	"goapptemp/config"
	"goapptemp/constant"
	"goapptemp/internal/adapter/repository"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	serror "goapptemp/internal/domain/service/error"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"strconv"
	"strings"
	"time"
)

var _ ClientService = (*clientService)(nil)

type ClientService interface {
	Create(ctx context.Context, req *CreateClientRequest) (*entity.Client, error)
	Update(ctx context.Context, req *UpdateClientRequest) (*entity.Client, error)
	Delete(ctx context.Context, req *DeleteClientRequest) error
	Find(ctx context.Context, req *FindClientsRequest) ([]*entity.Client, int, error)
	FindOne(ctx context.Context, req *FindOneClientRequest) (*entity.Client, error)
	IsDeletable(ctx context.Context, req *IsDeletableClientRequest) (bool, error)
}

type clientService struct {
	config *config.Config
	repo   repository.Repository
	log    logger.Logger
	auth   AuthService
	pubsub PubsubService
}

func NewClientService(config *config.Config, repo repository.Repository, log logger.Logger, auth AuthService, pubsub PubsubService) *clientService {
	return &clientService{
		config: config,
		repo:   repo,
		log:    log,
		auth:   auth,
		pubsub: pubsub,
	}
}

type CreateClientRequest struct {
	AuthParams *AuthParams
	Client     *entity.Client
}

func (s *clientService) Create(ctx context.Context, req *CreateClientRequest) (*entity.Client, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.CREATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.Client == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client data cannot be nil")
	}

	var createdClient *entity.Client

	var iconBase64, format string

	var isIconBase64 bool
	if req.Client.Icon != nil {
		isIconBase64 = true
		iconBase64 = *req.Client.Icon
		loadingStatus := constant.IconStatusLoading
		req.Client.Icon = &loadingStatus
		now := time.Now()
		req.Client.IconUpdatedAt = &now

		if len(iconBase64) >= constant.ImgMaxSize {
			return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Icon data too large")
		}

		format, err = shared.CheckBase64Image(iconBase64)
		if err != nil {
			return nil, exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Invalid icon format")
		}
	}

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		for {
			clientCode, err := shared.GenerateCode(constant.CodePefix["client"], 6)
			if err != nil {
				return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to generate client code")
			}

			exists, err := txRepo.Client().IsCodeExists(ctx, clientCode)
			if err != nil {
				return serror.TranslateRepoError(err)
			}

			if !exists {
				req.Client.Code = clientCode
				break
			}
		}

		createdClient, err = txRepo.Client().Create(ctx, req.Client)
		if err != nil {
			return err
		}

		req.Client.ID = createdClient.ID

		if len(req.Client.ClientSupportFeatures) > 0 {
			for i := range req.Client.ClientSupportFeatures {
				req.Client.ClientSupportFeatures[i].ClientID = createdClient.ID
			}

			_, err = txRepo.ClientSupportFeature().BulkCreate(ctx, req.Client.ClientSupportFeatures)
			if err != nil {
				return err
			}
		}

		if s.config.App.UsePubsub && isIconBase64 {
			fileName := strconv.FormatUint(uint64(createdClient.ID), 10) + "_" + time.Now().Format("20060102_150405") + "." + format
			userLog := strconv.FormatUint(uint64(req.AuthParams.AccessTokenClaims.UserID), 10)

			if err := s.pubsub.SendToPublisher(ctx, iconBase64, createdClient.ID, constant.ClientModelType, fileName, userLog); err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	createdClient.ClientSupportFeatures = nil

	return createdClient, nil
}

type FindClientsRequest struct {
	AuthParams *AuthParams
	Filter     *mysqlrepository.FilterClientPayload
}

func (s *clientService) Find(ctx context.Context, req *FindClientsRequest) ([]*entity.Client, int, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, 0, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.READ")
	if err != nil {
		return nil, 0, err
	}

	if !ok {
		return nil, 0, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	clients, totalCount, err := s.repo.MySQL().Client().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return clients, totalCount, nil
}

type FindOneClientRequest struct {
	AuthParams *AuthParams
	ClientID   uint
}

func (s *clientService) FindOne(ctx context.Context, req *FindOneClientRequest) (*entity.Client, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.READ")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.ClientID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client id cannot be zero")
	}

	client, err := s.repo.MySQL().Client().FindByID(ctx, req.ClientID, true)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return client, nil
}

type UpdateClientRequest struct {
	AuthParams *AuthParams
	Update     *mysqlrepository.UpdateClientPayload
}

func (s *clientService) Update(ctx context.Context, req *UpdateClientRequest) (*entity.Client, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.UPDATE")
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
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client ID required for update")
	}

	var updatedClient *entity.Client

	var iconBase64, format string

	var isIconBase64 bool

	if req.Update.Icon != nil {
		iconValue := *req.Update.Icon
		isURL := strings.HasPrefix(iconValue, "http://") || strings.HasPrefix(iconValue, "https://")

		isSpecialStatus := iconValue == constant.IconStatusLoading || iconValue == constant.IconStatusFailed
		if isSpecialStatus {
			req.Update.Icon = nil
		} else if !isURL {
			isIconBase64 = true

			iconBase64 = iconValue
			if len(iconBase64) >= constant.ImgMaxSize {
				return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Icon base64 data too large")
			}

			format, err = shared.CheckBase64Image(iconBase64)
			if err != nil {
				return nil, exception.Wrap(err, exception.TypeBadRequest, exception.CodeBadRequest, "Invalid icon base64 format")
			}

			loadingStatus := constant.IconStatusLoading
			req.Update.Icon = &loadingStatus
			now := time.Now()
			req.Update.IconUpdatedAt = &now
		}
	}

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		updatedClient, err = txRepo.Client().Update(ctx, req.Update)
		if err != nil {
			return err
		}

		err = txRepo.ClientSupportFeature().DeleteByClientID(ctx, updatedClient.ID)
		if err != nil {
			return err
		}

		if len(req.Update.ClientSupportFeatures) > 0 {
			_, err = txRepo.ClientSupportFeature().BulkCreate(ctx, req.Update.ClientSupportFeatures)
			if err != nil {
				return err
			}
		}

		if s.config.App.UsePubsub && isIconBase64 {
			fileName := strconv.FormatUint(uint64(updatedClient.ID), 10) + "_" + time.Now().Format("20060102_150405") + "." + format
			userLog := strconv.FormatUint(uint64(req.AuthParams.AccessTokenClaims.UserID), 10)

			if err = s.pubsub.SendToPublisher(ctx, iconBase64, updatedClient.ID, constant.ClientModelType, fileName, userLog); err != nil {
				return err
			}
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	updatedClient.ClientSupportFeatures = nil

	return updatedClient, nil
}

type DeleteClientRequest struct {
	AuthParams *AuthParams
	ClientID   uint
}

func (s *clientService) Delete(ctx context.Context, req *DeleteClientRequest) error {
	if req.AuthParams.AccessTokenClaims == nil {
		return exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.DELETE")
	if err != nil {
		return err
	}

	if !ok {
		return exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.ClientID == 0 {
		return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client ID cannot be zero")
	}

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		clientTable := txRepo.Client().GetTableName()
		ignoreTables := txRepo.ClientSupportFeature().GetTableName()

		dependencyMap, err := txRepo.StoreProcedure().CheckIfRecordsAreDeletable(ctx, clientTable, []uint{req.ClientID}, ignoreTables)
		if err != nil {
			return err
		}

		if count, found := dependencyMap[req.ClientID]; found && count > 0 {
			return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client is not deletable due to existing dependencies")
		}

		if err := txRepo.Client().Delete(ctx, req.ClientID); err != nil {
			return err
		}

		if err := txRepo.ClientSupportFeature().DeleteByClientID(ctx, req.ClientID); err != nil {
			return err
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return serror.TranslateRepoError(err)
	}

	return nil
}

type IsDeletableClientRequest struct {
	AuthParams *AuthParams
	ClientID   uint
}

func (s *clientService) IsDeletable(ctx context.Context, req *IsDeletableClientRequest) (bool, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return false, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "CLIENT.DELETE")
	if err != nil {
		return false, err
	}

	if !ok {
		return false, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.ClientID == 0 {
		return false, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Client ID cannot be zero")
	}

	clientTable := s.repo.MySQL().Client().GetTableName()
	ignoreTables := s.repo.MySQL().ClientSupportFeature().GetTableName()

	dependencyMap, err := s.repo.MySQL().StoreProcedure().CheckIfRecordsAreDeletable(ctx, clientTable, []uint{req.ClientID}, ignoreTables)
	if err != nil {
		return false, serror.TranslateRepoError(err)
	}

	if count, found := dependencyMap[req.ClientID]; found && count > 0 {
		return false, nil
	}

	return true, nil
}
