package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type City struct {
	bun.BaseModel `bun:"table:cities,alias:city"`
	ID            uint      `bun:"id,pk,autoincrement"`
	ProvinceID    uint      `bun:"province_id,notnull"`
	Province      *Province `bun:"rel:belongs-to,join:province_id=id"`
	Name          string    `bun:"name,notnull"`
}

func (m *City) ToDomain() *entity.City {
	if m == nil {
		return nil
	}

	res := &entity.City{
		ID:         m.ID,
		ProvinceID: m.ProvinceID,
		Name:       m.Name,
	}
	if m.Province != nil {
		res.Province = m.Province.ToDomain()
	}

	return res
}

func ToCitiesDomain(arg []*City) []*entity.City {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.City, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsCity(arg *entity.City) *City {
	if arg == nil {
		return nil
	}

	return &City{
		ID:         arg.ID,
		ProvinceID: arg.ProvinceID,
		Province:   AsProvince(arg.Province),
		Name:       arg.Name,
	}
}

func AsCities(arg []*entity.City) []*City {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*City, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsCity(arg[i]))
	}

	return res
}
