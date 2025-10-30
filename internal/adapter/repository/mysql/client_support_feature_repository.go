package mysqlrepository

import (
	"context"
	"database/sql"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/cockroachdb/errors"
	"github.com/uptrace/bun"
)

var _ ClientSupportFeatureRepository = (*clientSupportFeatureRepository)(nil)

type ClientSupportFeatureRepository interface {
	GetTableName() string
	BulkCreate(ctx context.Context, req []*entity.ClientSupportFeature) ([]*entity.ClientSupportFeature, error)
	DeleteByClientID(ctx context.Context, id uint) error
}

type clientSupportFeatureRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewClientSupportFeatureRepository(db bun.IDB, logger logger.Logger) *clientSupportFeatureRepository {
	return &clientSupportFeatureRepository{db: db, logger: logger}
}

func (r *clientSupportFeatureRepository) GetTableName() string {
	return "client_support_features"
}

func (r *clientSupportFeatureRepository) BulkCreate(ctx context.Context, req []*entity.ClientSupportFeature) ([]*entity.ClientSupportFeature, error) {
	if len(req) == 0 {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create client support feature")
	}

	clisfs := model.AsClientSupportFeatures(req)
	if _, err := r.db.NewInsert().Model(&clisfs).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create client support feature")
	}

	return model.ToClientSupportFeaturesDomain(clisfs), nil
}

func (r *clientSupportFeatureRepository) DeleteByClientID(ctx context.Context, clientID uint) error {
	if clientID == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete client support feature")
	}

	_, err := r.db.NewDelete().Model((*model.ClientSupportFeature)(nil)).Where("client_id = ?", clientID).Exec(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return handleDBError(err, r.GetTableName(), "delete client support feature")
	}

	return nil
}
