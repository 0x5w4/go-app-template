package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type UserRole struct {
	bun.BaseModel `bun:"table:user_roles,alias:usrrol"`
	UserID        uint  `bun:"user_id,pk"`
	User          *User `bun:"rel:belongs-to,join:user_id=id"`
	RoleID        uint  `bun:"role_id,pk"`
	Role          *Role `bun:"rel:belongs-to,join:role_id=id"`
}

func (m *UserRole) ToDomain() *entity.UserRole {
	if m == nil {
		return nil
	}

	return &entity.UserRole{
		UserID: m.UserID,
		RoleID: m.RoleID,
	}
}

func ToUserRolesDomain(arg []*UserRole) []*entity.UserRole {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.UserRole, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsUserRole(arg *entity.UserRole) *UserRole {
	if arg == nil {
		return nil
	}

	return &UserRole{
		UserID: arg.UserID,
		RoleID: arg.RoleID,
	}
}

func AsUserRoles(userID uint, roleIDs []uint) []*UserRole {
	if userID == 0 || len(roleIDs) == 0 {
		return nil
	}

	res := make([]*UserRole, 0, len(roleIDs))

	for i := range roleIDs {
		if roleIDs[i] == 0 {
			continue
		}

		res = append(res, AsUserRole(&entity.UserRole{UserID: userID, RoleID: roleIDs[i]}))
	}

	return res
}
