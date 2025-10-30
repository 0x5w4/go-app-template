package mysqlrepository

import (
	"context"
	"database/sql"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"

	"github.com/uptrace/bun"
)

var _ RoleRepository = (*roleRepository)(nil)

type RoleRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.Role) (*entity.Role, error)
	FindByID(ctx context.Context, id uint) (*entity.Role, error)
	Find(ctx context.Context, filter *FilterRolePayload) ([]*entity.Role, int, error)
	Update(ctx context.Context, req *UpdateRolePayload) (*entity.Role, error)
	Delete(ctx context.Context, id uint) error
	AttachPermissions(ctx context.Context, roleID uint, permissionIDs []uint) ([]*entity.RolePermission, error)
	DetachPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) ([]*entity.RolePermission, error)
}

type roleRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewRoleRepository(db bun.IDB, logger logger.Logger) *roleRepository {
	return &roleRepository{db: db, logger: logger}
}

func (r *roleRepository) GetTableName() string {
	return "roles"
}

func (r *roleRepository) Create(ctx context.Context, req *entity.Role) (*entity.Role, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create role")
	}

	role := model.AsRole(req)
	if _, err := r.db.NewInsert().Model(role).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create role")
	}

	return role.ToDomain(), nil
}

type FilterRolePayload struct {
	IDs        []uint
	Codes      []string
	Names      []string
	SuperAdmin *bool
	Search     string
	Page       int
	PerPage    int
}

func (r *roleRepository) Find(ctx context.Context, filter *FilterRolePayload) ([]*entity.Role, int, error) {
	var roles []*model.Role

	query := r.db.NewSelect().Model(&roles).Relation("Permissions")
	if len(filter.IDs) > 0 {
		query = query.Where("rol.id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.Codes) > 0 {
		query = query.Where("rol.code IN (?)", bun.In(filter.Codes))
	}

	if len(filter.Names) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Names {
				q = q.WhereOr("LOWER(rol.name) LIKE LOWER(?)", "%"+filter.Names[i]+"%")
			}

			return q
		})
	}

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(rol.code) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(rol.name) LIKE LOWER(?)", "%"+filter.Search+"%")

			return q
		})
	}

	if filter.SuperAdmin != nil {
		query = query.Where("rol.super_admin = ?", *filter.SuperAdmin)
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count role")
	}

	if totalCount == 0 {
		return []*entity.Role{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("rol.id DESC")
	if err = query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find role")
	}

	return model.ToRolesDomain(roles), totalCount, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uint) (*entity.Role, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find role by id")
	}

	role := &model.Role{Base: model.Base{ID: id}}
	if err := r.db.NewSelect().Model(role).Relation("Permissions").WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find role by id")
	}

	return role.ToDomain(), nil
}

type UpdateRolePayload struct {
	ID            uint
	PermissionIDs []*uint
	Code          *string
	Name          *string
	Description   *string
	SuperAdmin    *bool
}

func (r *roleRepository) Update(ctx context.Context, req *UpdateRolePayload) (*entity.Role, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update role")
	}

	current, err := r.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	role := model.AsRole(current)
	if req.Code != nil {
		role.Code = *req.Code
	}

	if req.Name != nil {
		role.Name = *req.Name
	}

	if req.Description != nil {
		role.Description = req.Description
	}

	if req.SuperAdmin != nil {
		role.SuperAdmin = *req.SuperAdmin
	}

	if _, err := r.db.NewUpdate().Model(role).WherePK().Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update role")
	}

	return role.ToDomain(), nil
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete role")
	}

	role := &model.Role{Base: model.Base{ID: id}}

	res, err := r.db.NewDelete().Model(role).WherePK().Exec(ctx)
	if err != nil {
		return handleDBError(err, r.GetTableName(), "delete role")
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return handleDBError(sql.ErrNoRows, r.GetTableName(), "delete role")
	}

	return nil
}

func (r *roleRepository) AttachPermissions(ctx context.Context, roleID uint, permissionIDs []uint) ([]*entity.RolePermission, error) {
	if roleID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "attach permissions to role")
	}

	if len(permissionIDs) == 0 {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "attach permissions to role")
	}

	rolePermissions := model.AsRolePermissions(roleID, permissionIDs)
	if _, err := r.db.NewInsert().Model(&rolePermissions).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "attach permissions to role")
	}

	return model.ToRolePermissionsDomain(rolePermissions), nil
}

func (r *roleRepository) DetachPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	if roleID == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "detach permissions from role")
	}

	if len(permissionIDs) == 0 {
		return handleDBError(exception.ErrDataNull, r.GetTableName(), "detach permissions from role")
	}

	rolePermissions := model.AsRolePermissions(roleID, permissionIDs)

	res, err := r.db.NewDelete().Model(&rolePermissions).WherePK().Exec(ctx)
	if err != nil {
		return handleDBError(err, r.GetTableName(), "detach permissions from role")
	}

	_, err = res.RowsAffected()
	if err != nil {
		return handleDBError(sql.ErrNoRows, r.GetTableName(), "detach permissions from role")
	}

	return nil
}

func (r *roleRepository) SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) ([]*entity.RolePermission, error) {
	if roleID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "sync permissions to role")
	}

	role, err := r.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if len(role.PermissionIDs) != 0 {
		if err := r.DetachPermissions(ctx, roleID, role.PermissionIDs); err != nil {
			return nil, err
		}
	}

	var rolePermissions []*entity.RolePermission
	if len(permissionIDs) != 0 {
		rolePermissions, err = r.AttachPermissions(ctx, roleID, permissionIDs)
		if err != nil {
			return nil, err
		}
	}

	return rolePermissions, nil
}
