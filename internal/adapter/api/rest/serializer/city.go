package serializer

import (
	"goapptemp/internal/domain/entity"
)

type CityResponseData struct {
	ID         uint                  `json:"id"`
	ProvinceID uint                  `json:"province_id,omitempty"`
	Province   *ProvinceResponseData `json:"province,omitempty"`
	Name       string                `json:"name,omitempty"`
}

func SerializeCity(arg *entity.City) *CityResponseData {
	if arg == nil {
		return nil
	}

	return &CityResponseData{
		ID:         arg.ID,
		ProvinceID: arg.ProvinceID,
		Province:   SerializeProvince(arg.Province),
		Name:       arg.Name,
	}
}

func SerializeCities(arg []*entity.City) []*CityResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*CityResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeCity(arg[i]))
	}

	return res
}
