package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type Permission struct {
	bun.BaseModel `bun:"table:permissions,alias:perm"`
	Base
	Roles       []*Role `bun:"m2m:role_permissions,join:Permission=Role"`
	Code        string  `bun:"code,notnull"`
	Name        string  `bun:"name,notnull"`
	Description *string `bun:"description"`
	CodeActive  *string `bun:"code_active,scanonly"`
	NameActive  *string `bun:"name_active,scanonly"`
}

func (m *Permission) ToDomain() *entity.Permission {
	if m == nil {
		return nil
	}

	return &entity.Permission{
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
}

func ToPermissionsDomain(arg []*Permission) []*entity.Permission {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.Permission, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsPermission(arg *entity.Permission) *Permission {
	if arg == nil {
		return nil
	}

	return &Permission{
		Code:        arg.Code,
		Name:        arg.Name,
		Description: arg.Description,
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsPermissions(arg []*entity.Permission) []*Permission {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*Permission, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsPermission(arg[i]))
	}

	return res
}
