package mysqlrepository

import (
	"context"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ PermissionRepository = (*permissionRepository)(nil)

type PermissionRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.Permission) (*entity.Permission, error)
	FindByID(ctx context.Context, id uint) (*entity.Permission, error)
	Find(ctx context.Context, filter *FilterPermissionPayload) ([]*entity.Permission, int, error)
	Update(ctx context.Context, req *UpdatePermissionPayload) (*entity.Permission, error)
	Delete(ctx context.Context, id uint) error
}

type permissionRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewPermissionRepository(db bun.IDB, logger logger.Logger) *permissionRepository {
	return &permissionRepository{db: db, logger: logger}
}

func (r *permissionRepository) GetTableName() string {
	return "permissions"
}

func (r *permissionRepository) Create(ctx context.Context, req *entity.Permission) (*entity.Permission, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create permission")
	}

	permission := model.AsPermission(req)
	if _, err := r.db.NewInsert().Model(permission).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create permission")
	}

	return permission.ToDomain(), nil
}

func (r *permissionRepository) FindByID(ctx context.Context, id uint) (*entity.Permission, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find permission by id")
	}

	permission := &model.Permission{Base: model.Base{ID: id}}
	if err := r.db.NewSelect().Model(permission).WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find permission by id")
	}

	return permission.ToDomain(), nil
}

type FilterPermissionPayload struct {
	IDs     []uint
	Codes   []string
	Names   []string
	Search  string
	Page    int
	PerPage int
}

func (r *permissionRepository) Find(ctx context.Context, filter *FilterPermissionPayload) ([]*entity.Permission, int, error) {
	var permissions []*model.Permission

	query := r.db.NewSelect().Model(&permissions)
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.Codes) > 0 {
		query = query.Where("code IN (?)", bun.In(filter.Codes))
	}

	if len(filter.Names) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Names {
				q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Names[i]+"%")
			}

			return q
		})
	}

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(code) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(name) LIKE LOWER(?)", "%"+filter.Search+"%")

			return q
		})
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count permission")
	}

	if totalCount == 0 {
		return []*entity.Permission{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("id ASC")
	if err = query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find permission")
	}

	return model.ToPermissionsDomain(permissions), totalCount, nil
}

type UpdatePermissionPayload struct {
	ID          uint
	Code        *string
	Name        *string
	Description *string
}

func (r *permissionRepository) Update(ctx context.Context, req *UpdatePermissionPayload) (*entity.Permission, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update permission")
	}

	permission := &model.Permission{Base: model.Base{ID: req.ID}}

	var columnsToUpdate []string

	if req.Name != nil {
		permission.Name = *req.Name

		columnsToUpdate = append(columnsToUpdate, "name")
	}

	if req.Description != nil {
		permission.Description = req.Description

		columnsToUpdate = append(columnsToUpdate, "description")
	}

	if len(columnsToUpdate) == 0 {
		return permission.ToDomain(), nil
	}

	query := r.db.NewUpdate().Model(permission).Column(columnsToUpdate...).WherePK().Returning("*")
	if _, err := query.Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update principle")
	}

	return permission.ToDomain(), nil
}

func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete permission")
	}

	permission := &model.Permission{Base: model.Base{ID: id}}
	if _, err := r.db.NewDelete().Model(permission).WherePK().Exec(ctx); err != nil {
		return handleDBError(err, r.GetTableName(), "delete permission")
	}

	return nil
}
