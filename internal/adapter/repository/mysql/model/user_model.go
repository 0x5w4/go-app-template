package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:usr"`
	Base
	Roles          []*Role `bun:"m2m:user_roles,join:User=Role"`
	CompanyID      uint
	Company        *Company `bun:"rel:belongs-to,join:company_id=id"`
	Username       string   `bun:"username,notnull"`
	Email          string   `bun:"email,notnull"`
	Password       string   `bun:"password,notnull"`
	Fullname       string   `bun:"fullname,notnull"`
	UsernameActive *string  `bun:"username_active,unique:uq_users_company_username_active"`
	EmailActive    *string  `bun:"email_active,unique:uq_users_company_email_active"`
}

func (m *User) ToDomain() *entity.User {
	if m == nil {
		return nil
	}

	roleIDs := make([]uint, 0, len(m.Roles))

	for i := range m.Roles {
		if m.Roles[i] == nil {
			continue
		}

		roleIDs = append(roleIDs, m.Roles[i].ID)
	}

	return &entity.User{
		RoleIDs:   roleIDs,
		Roles:     ToRolesDomain(m.Roles),
		CompanyID: m.CompanyID,
		Username:  m.Username,
		Email:     m.Email,
		Password:  m.Password,
		Fullname:  m.Fullname,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
}

func ToUsersDomain(arg []*User) []*entity.User {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.User, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsUser(arg *entity.User) *User {
	if arg == nil {
		return nil
	}

	return &User{
		Roles:     AsRoles(arg.Roles),
		CompanyID: arg.CompanyID,
		Username:  arg.Username,
		Email:     arg.Email,
		Password:  arg.Password,
		Fullname:  arg.Fullname,
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsUsers(arg []*entity.User) []*User {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*User, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsUser(arg[i]))
	}

	return res
}
