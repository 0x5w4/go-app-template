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

var _ UserRepository = (*userRepository)(nil)

type UserRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.User) (*entity.User, error)
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	Find(ctx context.Context, filter *FilterUserPayload) ([]*entity.User, int, error)
	Update(ctx context.Context, req *UpdateUserPayload) (*entity.User, error)
	Delete(ctx context.Context, id uint) error
	AttachRoles(ctx context.Context, userID uint, roleIDs []uint) ([]*entity.UserRole, error)
	DetachRoles(ctx context.Context, userID uint, roleIDs []uint) error
	SyncRoles(ctx context.Context, userID uint, roleIDs []uint) ([]*entity.UserRole, error)
	HasPermission(ctx context.Context, userID uint, permissionCode string) (bool, error)
}

type userRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewUserRepository(db bun.IDB, logger logger.Logger) *userRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) GetTableName() string {
	return "users"
}

func (r *userRepository) Create(ctx context.Context, req *entity.User) (*entity.User, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create user")
	}

	user := model.AsUser(req)
	if _, err := r.db.NewInsert().Model(user).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create user")
	}

	return user.ToDomain(), nil
}

type FilterUserPayload struct {
	IDs       []uint
	Fullnames []string
	Usernames []string
	Emails    []string
	Search    string
	Page      int
	PerPage   int
}

func (r *userRepository) Find(ctx context.Context, filter *FilterUserPayload) ([]*entity.User, int, error) {
	var users []*model.User

	query := r.db.NewSelect().Model(&users).Relation("Roles.Permissions")
	if len(filter.IDs) > 0 {
		query = query.Where("usr.id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.Emails) > 0 {
		query = query.Where("usr.email IN (?)", bun.In(filter.Emails))
	}

	if len(filter.Usernames) > 0 {
		query = query.Where("usr.username IN (?)", bun.In(filter.Usernames))
	}

	if len(filter.Fullnames) > 0 {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i := range filter.Fullnames {
				q = q.WhereOr("LOWER(usr.fullname) LIKE LOWER(?)", "%"+filter.Fullnames[i]+"%")
			}

			return q
		})
	}

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(usr.email) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(usr.username) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(usr.fullname) LIKE LOWER(?)", "%"+filter.Search+"%")

			return q
		})
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count user")
	}

	if totalCount == 0 {
		return []*entity.User{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("usr.id DESC")
	if err = query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find user")
	}

	return model.ToUsersDomain(users), totalCount, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find user by id")
	}

	user := &model.User{Base: model.Base{ID: id}}
	if err := r.db.NewSelect().Model(user).Relation("Roles.Permissions").WherePK().Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "find user by id")
	}

	return user.ToDomain(), nil
}

type UpdateUserPayload struct {
	ID       uint
	RoleIDs  []*uint
	Fullname *string
	Username *string
	Email    *string
	Password *string
}

func (r *userRepository) Update(ctx context.Context, req *UpdateUserPayload) (*entity.User, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update user: ID is zero")
	}

	userModel := &model.User{
		Base: model.Base{ID: req.ID},
	}

	var columnsToUpdate []string

	if req.Fullname != nil {
		userModel.Fullname = *req.Fullname

		columnsToUpdate = append(columnsToUpdate, "fullname")
	}

	if req.Username != nil {
		userModel.Username = *req.Username

		columnsToUpdate = append(columnsToUpdate, "username")
	}

	if req.Email != nil {
		userModel.Email = *req.Email

		columnsToUpdate = append(columnsToUpdate, "email")
	}

	if req.Password != nil && *req.Password != "" {
		userModel.Password = *req.Password

		columnsToUpdate = append(columnsToUpdate, "password")
	}

	if len(columnsToUpdate) == 0 {
		currentUser, err := r.FindByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}

		return currentUser, nil
	}

	query := r.db.NewUpdate().
		Model(userModel).
		Column(columnsToUpdate...).
		WherePK()
	if _, err := query.Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update user")
	}

	return userModel.ToDomain(), nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete user")
	}

	user := &model.User{Base: model.Base{ID: id}}

	res, err := r.db.NewDelete().Model(user).WherePK().Exec(ctx)
	if err != nil {
		return handleDBError(err, r.GetTableName(), "delete user")
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return handleDBError(sql.ErrNoRows, r.GetTableName(), "delete user")
	}

	return nil
}

func (r *userRepository) AttachRoles(ctx context.Context, userID uint, roleIDs []uint) ([]*entity.UserRole, error) {
	if userID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "attach roles to user")
	}

	if len(roleIDs) == 0 {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "attach roles to user")
	}

	userRoles := model.AsUserRoles(userID, roleIDs)
	if _, err := r.db.NewInsert().Model(&userRoles).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "attach roles to user")
	}

	return model.ToUserRolesDomain(userRoles), nil
}

func (r *userRepository) DetachRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	if userID == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "detach roles from user")
	}

	if len(roleIDs) == 0 {
		return handleDBError(exception.ErrDataNull, r.GetTableName(), "detach roles from user")
	}

	userRoles := model.AsUserRoles(userID, roleIDs)

	res, err := r.db.NewDelete().Model(&userRoles).WherePK().Exec(ctx)
	if err != nil {
		return handleDBError(err, r.GetTableName(), "detach roles from user")
	}

	_, err = res.RowsAffected()
	if err != nil {
		return handleDBError(err, r.GetTableName(), "detach roles from user")
	}

	return nil
}

func (r *userRepository) SyncRoles(ctx context.Context, userID uint, roleIDs []uint) ([]*entity.UserRole, error) {
	if userID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "sync roles to user")
	}

	var userRoles []*entity.UserRole

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var currentRoleIDs []uint

		err := tx.NewSelect().
			Model((*model.UserRole)(nil)).
			Column("role_id").
			Where("user_id = ?", userID).
			Scan(ctx, &currentRoleIDs)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return handleDBError(err, r.GetTableName(), "sync roles: get current roles")
		}

		if len(currentRoleIDs) > 0 {
			userRolesToDetach := model.AsUserRoles(userID, currentRoleIDs)
			if _, err := tx.NewDelete().Model(&userRolesToDetach).WherePK().Exec(ctx); err != nil {
				return handleDBError(err, r.GetTableName(), "sync roles: detach")
			}
		}

		if len(roleIDs) > 0 {
			userRolesToAttach := model.AsUserRoles(userID, roleIDs)

			if _, err := tx.NewInsert().Model(&userRolesToAttach).Returning("*").Exec(ctx, &userRolesToAttach); err != nil {
				return handleDBError(err, r.GetTableName(), "sync roles: attach")
			}

			userRoles = model.ToUserRolesDomain(userRolesToAttach)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to sync roles in transaction")
	}

	return userRoles, nil
}

func (r *userRepository) HasPermission(ctx context.Context, userID uint, permissionCode string) (bool, error) {
	if userID == 0 {
		return false, handleDBError(exception.ErrIDNull, r.GetTableName(), "check permission")
	}

	if permissionCode == "" {
		return false, handleDBError(exception.ErrDataNull, r.GetTableName(), "check permission: permission code is empty")
	}

	superAdminQuery := r.db.NewSelect().
		Model((*model.UserRole)(nil)).
		Join("JOIN ? AS r ON r.id = user_role.role_id", r.db.NewSelect().Model((*model.Role)(nil))).
		Where("user_role.user_id = ?", userID).
		Where("r.super_admin = ?", true)

	permissionQuery := r.db.NewSelect().
		Model((*model.UserRole)(nil)).
		Join("JOIN ? AS rp ON rp.role_id = user_role.role_id", r.db.NewSelect().Model((*model.RolePermission)(nil))).
		Join("JOIN ? AS p ON p.id = rp.permission_id", r.db.NewSelect().Model((*model.Permission)(nil))).
		Where("user_role.user_id = ?", userID).
		Where("p.code = ?", permissionCode)

	hasPermission, err := r.db.NewSelect().
		TableExpr("(?) AS combined_query", superAdminQuery.Union(permissionQuery)).
		Exists(ctx)
	if err != nil {
		return false, handleDBError(err, r.GetTableName(), "check permission")
	}

	return hasPermission, nil
}
