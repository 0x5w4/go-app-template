package model

import (
	"goapptemp/internal/domain/entity"
	"time"

	"github.com/uptrace/bun"
)

type Company struct {
	bun.BaseModel `bun:"table:companies,alias:comp"`
	Base
	AdminID       uint       `bun:"admin_id,notnull"`
	Name          string     `bun:"name,notnull"`
	Icon          *string    `bun:"icon"`
	IconUpdatedAt *time.Time `bun:"icon_updated_at"`
	NameActive    *string    `bun:"name_active,unique:uq_companies_name_active"`
}

func (m *Company) ToDomain() *entity.Company {
	if m == nil {
		return nil
	}

	res := &entity.Company{
		Name:          m.Name,
		Icon:          m.Icon,
		IconUpdatedAt: m.IconUpdatedAt,
		AdminID:       m.AdminID,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}

	return res
}

func ToCompaniesDomain(arg []*Company) []*entity.Company {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.Company, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsCompany(arg *entity.Company) *Company {
	if arg == nil {
		return nil
	}

	return &Company{
		Name:          arg.Name,
		Icon:          arg.Icon,
		IconUpdatedAt: arg.IconUpdatedAt,
		AdminID:       arg.AdminID,
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsCompanies(arg []*entity.Company) []*Company {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*Company, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsCompany(arg[i]))
	}

	return res
}
