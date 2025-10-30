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

var _ SupportFeatureRepository = (*supportFeatureRepository)(nil)

type SupportFeatureRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.SupportFeature) (*entity.SupportFeature, error)
	BulkCreate(ctx context.Context, req []*entity.SupportFeature) ([]*entity.SupportFeature, error)
	FindByID(ctx context.Context, id uint) (*entity.SupportFeature, error)
	Find(ctx context.Context, filter *FilterSupportFeaturePayload) ([]*entity.SupportFeature, int, error)
	Update(ctx context.Context, req *UpdateSupportFeaturePayload) (*entity.SupportFeature, error)
	Delete(ctx context.Context, id uint) error
	GetExistingCodes(ctx context.Context, codes []string) (map[string]struct{}, error)
	CheckCodeExists(ctx context.Context, code string) (bool, error)
	CheckKeyExists(ctx context.Context, key string, id *uint) (bool, error)
	FindExistingKeysAndNames(ctx context.Context, keys []string, names []string) (existingKeys map[string]struct{}, existingNames map[string]struct{}, err error)
}

type supportFeatureRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewSupportFeatureRepository(db bun.IDB, logger logger.Logger) *supportFeatureRepository {
	return &supportFeatureRepository{db: db, logger: logger}
}

func (r *supportFeatureRepository) GetTableName() string {
	return "support_features"
}

func (r *supportFeatureRepository) Create(ctx context.Context, req *entity.SupportFeature) (*entity.SupportFeature, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create support feature")
	}

	supportFeature := model.AsSupportFeature(req)
	if _, err := r.db.NewInsert().Model(supportFeature).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create support feature")
	}

	return supportFeature.ToDomain(), nil
}

func (r *supportFeatureRepository) BulkCreate(ctx context.Context, req []*entity.SupportFeature) ([]*entity.SupportFeature, error) {
	if len(req) == 0 {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "bulk create support features")
	}

	supportFeatures := model.AsSupportFeatures(req)
	if _, err := r.db.NewInsert().Model(&supportFeatures).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "bulk create support features")
	}

	return model.ToSupportFeaturesDomain(supportFeatures), nil
}

type FilterSupportFeaturePayload struct {
	IDs      []uint
	Codes    []string
	Names    []string
	Keys     []string
	IsActive *bool
	Search   string
	Page     int
	PerPage  int
}

func (r *supportFeatureRepository) Find(ctx context.Context, filter *FilterSupportFeaturePayload) ([]*entity.SupportFeature, int, error) {
	var supportFeatures []*model.SupportFeature

	query := r.db.NewSelect().Model(&supportFeatures)
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.Codes) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Codes {
				q = q.WhereOr("LOWER(code) LIKE LOWER(?)", "%"+filter.Codes[i]+"%")
			}

			return q
		})
	}

	if len(filter.Names) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Names {
				q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Names[i]+"%")
			}

			return q
		})
	}

	if len(filter.Keys) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Keys {
				q = q.WhereOr("LOWER(`key`) LIKE LOWER(?)", "%"+filter.Keys[i]+"%")
			}

			return q
		})
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(code) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(`key`) LIKE LOWER(?)", "%"+filter.Search+"%")

			return q
		})
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count support feature")
	}

	if totalCount == 0 {
		return []*entity.SupportFeature{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("id DESC")
	if err = query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find support feature")
	}

	return model.ToSupportFeaturesDomain(supportFeatures), totalCount, nil
}

func (r *supportFeatureRepository) FindByID(ctx context.Context, id uint) (*entity.SupportFeature, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find support feature by id")
	}

	supportFeature := &model.SupportFeature{Base: model.Base{ID: id}}
	if err := r.db.NewSelect().Model(supportFeature).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find support feature by id")
	}

	return supportFeature.ToDomain(), nil
}

type UpdateSupportFeaturePayload struct {
	ID       uint
	Code     *string
	Name     *string
	Key      *string
	IsActive *bool
}

func (r *supportFeatureRepository) Update(ctx context.Context, req *UpdateSupportFeaturePayload) (*entity.SupportFeature, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update support feature: ID is zero")
	}

	supportFeatureModel := &model.SupportFeature{
		Base: model.Base{ID: req.ID},
	}

	var columnsToUpdate []string

	if req.Code != nil {
		supportFeatureModel.Code = *req.Code

		columnsToUpdate = append(columnsToUpdate, "code")
	}

	if req.Name != nil {
		supportFeatureModel.Name = *req.Name

		columnsToUpdate = append(columnsToUpdate, "name")
	}

	if req.Key != nil {
		supportFeatureModel.Key = *req.Key

		columnsToUpdate = append(columnsToUpdate, "key")
	}

	if req.IsActive != nil {
		supportFeatureModel.IsActive = *req.IsActive

		columnsToUpdate = append(columnsToUpdate, "is_active")
	}

	if len(columnsToUpdate) == 0 {
		currentSupportFeature, err := r.FindByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}

		return currentSupportFeature, nil
	}

	query := r.db.NewUpdate().Model(supportFeatureModel).Column(columnsToUpdate...).WherePK()
	if _, err := query.Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update support feature")
	}

	return supportFeatureModel.ToDomain(), nil
}

func (r *supportFeatureRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete support feature")
	}

	supportFeature := &model.SupportFeature{Base: model.Base{ID: id}}
	if _, err := r.db.NewDelete().Model(supportFeature).WherePK().Exec(ctx); err != nil {
		return handleDBError(err, r.GetTableName(), "delete support feature")
	}

	return nil
}

func (r *supportFeatureRepository) GetExistingCodes(ctx context.Context, codes []string) (map[string]struct{}, error) {
	if len(codes) == 0 {
		return make(map[string]struct{}), nil
	}

	var existingDbCodes []string

	err := r.db.NewSelect().
		Model((*model.SupportFeature)(nil)).
		Column("code").
		Where("code_active IN (?)", bun.In(codes)).
		Scan(ctx, &existingDbCodes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(map[string]struct{}), nil
		}

		return nil, handleDBError(err, r.GetTableName(), "get existing support feature codes")
	}

	resultSet := make(map[string]struct{})
	for _, code := range existingDbCodes {
		resultSet[code] = struct{}{}
	}

	return resultSet, nil
}

func (r *supportFeatureRepository) CheckCodeExists(ctx context.Context, code string) (bool, error) {
	if code == "" {
		return false, handleDBError(exception.ErrDataNull, r.GetTableName(), "check support feature code exists")
	}

	query := r.db.NewSelect().Model((*model.SupportFeature)(nil)).Where("code_active = ?", code)

	exist, err := query.Exists(ctx)
	if err != nil {
		return false, handleDBError(err, r.GetTableName(), "check support feature code exists")
	}

	return exist, nil
}

func (r *supportFeatureRepository) CheckKeyExists(ctx context.Context, key string, id *uint) (bool, error) {
	if key == "" {
		return false, handleDBError(exception.ErrDataNull, r.GetTableName(), "check support feature key exists")
	}

	query := r.db.NewSelect().Model((*model.SupportFeature)(nil)).Where("key_active = ?", key)
	if id != nil && *id > 0 {
		query = query.Where("id != ?", *id)
	}

	exist, err := query.Exists(ctx)
	if err != nil {
		return false, handleDBError(err, r.GetTableName(), "check support feature key exists")
	}

	return exist, nil
}

func (r *supportFeatureRepository) FindExistingKeysAndNames(ctx context.Context, keys []string, names []string) (map[string]struct{}, map[string]struct{}, error) {
	existingKeys := make(map[string]struct{})
	existingNames := make(map[string]struct{})

	if len(keys) == 0 && len(names) == 0 {
		return existingKeys, existingNames, nil
	}

	var results []struct {
		Key  string `bun:"key"`
		Name string `bun:"name"`
	}

	query := r.db.NewSelect().Model((*model.SupportFeature)(nil)).Column("key", "name")
	query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		for _, name := range names {
			q = q.WhereOr("name_active = ?", name)
		}

		for _, key := range keys {
			q = q.WhereOr("key_active = ?", key)
		}

		return q
	})

	if err := query.Scan(ctx, &results); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return existingKeys, existingNames, nil
		}

		return nil, nil, handleDBError(err, r.GetTableName(), "find existing keys and names")
	}

	for _, result := range results {
		if result.Key != "" {
			existingKeys[result.Key] = struct{}{}
		}

		if result.Name != "" {
			existingNames[result.Name] = struct{}{}
		}
	}

	return existingKeys, existingNames, nil
}
