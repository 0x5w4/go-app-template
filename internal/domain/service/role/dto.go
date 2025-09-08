package role

import (
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/auth"
)

type CreateRoleRequest struct {
	AuthParams *auth.AuthParams
	Role       *entity.Role
}

type UpdateRoleRequest struct {
	AuthParams *auth.AuthParams
	Update     *mysql.UpdateRolePayload
}

type DeleteRoleRequest struct {
	AuthParams *auth.AuthParams
	RoleID     uint
}

type FindRolesRequest struct {
	AuthParams *auth.AuthParams
	Filter     *mysql.FilterRolePayload
}

type FindOneRoleRequest struct {
	AuthParams *auth.AuthParams
	RoleID     uint
}
