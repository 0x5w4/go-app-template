package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:rol"`
	Base
	Permissions []*Permission `bun:"m2m:role_permissions,join:Role=Permission"`
	Code        string        `bun:"code,notnull"`
	Name        string        `bun:"name,notnull"`
	Description *string       `bun:"description"`
	SuperAdmin  bool          `bun:"super_admin,notnull"`
	CodeActive  *string       `bun:"code_active,unique:uq_roles_code_active"`
	NameActive  *string       `bun:"name_active,unique:uq_roles_name_active"`
}

func (m *Role) ToDomain() *entity.Role {
	if m == nil {
		return nil
	}

	permissionIDs := make([]uint, 0, len(m.Permissions))

	for i := range m.Permissions {
		if m.Permissions[i] == nil {
			continue
		}

		permissionIDs = append(permissionIDs, m.Permissions[i].ID)
	}

	res := &entity.Role{
		PermissionIDs: permissionIDs,
		Permissions:   ToPermissionsDomain(m.Permissions),
		Code:          m.Code,
		Name:          m.Name,
		Description:   m.Description,
		SuperAdmin:    m.SuperAdmin,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}

	return res
}

func ToRolesDomain(arg []*Role) []*entity.Role {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.Role, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsRole(arg *entity.Role) *Role {
	if arg == nil {
		return nil
	}

	return &Role{
		Permissions: AsPermissions(arg.Permissions),
		Code:        arg.Code,
		Name:        arg.Name,
		Description: arg.Description,
		SuperAdmin:  arg.SuperAdmin,
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsRoles(arg []*entity.Role) []*Role {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*Role, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsRole(arg[i]))
	}

	return res
}
