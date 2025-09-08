package serializer

import (
	"goapptemp/internal/domain/entity"
)

type DistrictResponseData struct {
	ID     uint              `json:"id"`
	CityID uint              `json:"city_id"`
	City   *CityResponseData `json:"city,omitempty"`
	Name   string            `json:"name"`
}

func SerializeDistrict(arg *entity.District) *DistrictResponseData {
	if arg == nil {
		return nil
	}

	return &DistrictResponseData{
		ID:     arg.ID,
		CityID: arg.CityID,
		City:   SerializeCity(arg.City),
		Name:   arg.Name,
	}
}

func SerializeDistricts(arg []*entity.District) []*DistrictResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*DistrictResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeDistrict(arg[i]))
	}

	return res
}
