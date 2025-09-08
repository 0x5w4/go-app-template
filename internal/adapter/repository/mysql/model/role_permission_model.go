package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type RolePermission struct {
	bun.BaseModel `bun:"table:role_permissions,alias:rolperm"`
	RoleId        uint        `bun:"role_id,pk"`
	Role          *Role       `bun:"rel:belongs-to,join:role_id=id"`
	PermissionId  uint        `bun:"permission_id,pk"`
	Permission    *Permission `bun:"rel:belongs-to,join:permission_id=id"`
}

func (m *RolePermission) ToDomain() *entity.RolePermission {
	if m == nil {
		return nil
	}

	res := &entity.RolePermission{
		RoleId:       m.RoleId,
		PermissionId: m.PermissionId,
	}
	if m.Role != nil {
		res.Role = m.Role.ToDomain()
	}

	if m.Permission != nil {
		res.Permission = m.Permission.ToDomain()
	}

	return res
}

func ToRolePermissionsDomain(arg []*RolePermission) []*entity.RolePermission {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.RolePermission, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsRolePermission(arg *entity.RolePermission) *RolePermission {
	if arg == nil {
		return nil
	}

	return &RolePermission{
		RoleId:       arg.RoleId,
		Role:         AsRole(arg.Role),
		PermissionId: arg.PermissionId,
		Permission:   AsPermission(arg.Permission),
	}
}

func AsRolePermissions(roleID uint, permissionIDs []uint) []*RolePermission {
	if roleID == 0 || len(permissionIDs) == 0 {
		return nil
	}

	res := make([]*RolePermission, 0, len(permissionIDs))

	for i := range permissionIDs {
		if permissionIDs[i] == 0 {
			continue
		}

		res = append(res, AsRolePermission(&entity.RolePermission{RoleId: roleID, PermissionId: permissionIDs[i]}))
	}

	return res
}
