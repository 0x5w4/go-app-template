package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type Province struct {
	bun.BaseModel `bun:"table:provinces,alias:prov"`
	ID            uint   `bun:"id,pk,autoincrement"`
	Name          string `bun:"name,notnull"`
}

func (m *Province) ToDomain() *entity.Province {
	if m == nil {
		return nil
	}

	return &entity.Province{
		ID:   m.ID,
		Name: m.Name,
	}
}

func ToProvincesDomain(arg []*Province) []*entity.Province {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.Province, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsProvince(arg *entity.Province) *Province {
	if arg == nil {
		return nil
	}

	return &Province{
		ID:   arg.ID,
		Name: arg.Name,
	}
}

func AsProvinces(arg []*entity.Province) []*Province {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*Province, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsProvince(arg[i]))
	}

	return res
}
