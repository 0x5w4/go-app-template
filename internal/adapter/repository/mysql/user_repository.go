package mysql

import (
	"context"
	"database/sql"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/pkg/logger"

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
		return nil, handleDBError(ErrDataNull, r.GetTableName(), "create user")
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
		return nil, handleDBError(ErrIDNull, r.GetTableName(), "find user by id")
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
		return nil, handleDBError(ErrIDNull, r.GetTableName(), "update user: ID is zero")
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
		return handleDBError(ErrIDNull, r.GetTableName(), "delete user")
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
		return nil, handleDBError(ErrIDNull, r.GetTableName(), "attach roles to user")
	}

	if len(roleIDs) == 0 {
		return nil, handleDBError(ErrDataNull, r.GetTableName(), "attach roles to user")
	}

	userRoles := model.AsUserRoles(userID, roleIDs)
	if _, err := r.db.NewInsert().Model(&userRoles).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "attach roles to user")
	}

	return model.ToUserRolesDomain(userRoles), nil
}

func (r *userRepository) DetachRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	if userID == 0 {
		return handleDBError(ErrIDNull, r.GetTableName(), "detach roles from user")
	}

	if len(roleIDs) == 0 {
		return handleDBError(ErrDataNull, r.GetTableName(), "detach roles from user")
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
		return nil, handleDBError(ErrIDNull, r.GetTableName(), "sync roles to user")
	}

	user, err := r.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(user.RoleIDs) != 0 {
		if err := r.DetachRoles(ctx, userID, user.RoleIDs); err != nil {
			return nil, err
		}
	}

	var userRoles []*entity.UserRole
	if len(roleIDs) != 0 {
		userRoles, err = r.AttachRoles(ctx, userID, roleIDs)
		if err != nil {
			return nil, err
		}
	}

	return userRoles, nil
}
