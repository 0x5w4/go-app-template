package user

import (
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/auth"
)

type CreateUserRequest struct {
	AuthParams *auth.AuthParams
	User       *entity.User
}

type UpdateUserRequest struct {
	AuthParams *auth.AuthParams
	Update     *mysql.UpdateUserPayload
}

type DeleteUserRequest struct {
	AuthParams *auth.AuthParams
	UserID     uint
}

type FindUserRequest struct {
	AuthParams *auth.AuthParams
	UserFilter *mysql.FilterUserPayload
}

type FindOneUserRequest struct {
	AuthParams *auth.AuthParams
	UserID     uint
}
