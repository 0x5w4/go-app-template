package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type District struct {
	bun.BaseModel `bun:"table:districts,alias:dist"`
	ID            uint   `bun:"id,pk,autoincrement"`
	CityID        uint   `bun:"city_id,notnull"`
	City          *City  `bun:"rel:belongs-to,join:city_id=id"`
	Name          string `bun:"name,notnull"`
}

func (m *District) ToDomain() *entity.District {
	if m == nil {
		return nil
	}

	res := &entity.District{
		ID:     m.ID,
		CityID: m.CityID,
		Name:   m.Name,
	}
	if m.City != nil {
		res.City = m.City.ToDomain()
	}

	return res
}

func ToDistrictsDomain(arg []*District) []*entity.District {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.District, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsDistrict(arg *entity.District) *District {
	if arg == nil {
		return nil
	}

	return &District{
		ID:     arg.ID,
		CityID: arg.CityID,
		City:   AsCity(arg.City),
		Name:   arg.Name,
	}
}

func AsDistricts(arg []*entity.District) []*District {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*District, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsDistrict(arg[i]))
	}

	return res
}
