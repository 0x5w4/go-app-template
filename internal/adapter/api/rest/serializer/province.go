package serializer

import (
	"goapptemp/internal/domain/entity"
)

type ProvinceResponseData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func SerializeProvince(arg *entity.Province) *ProvinceResponseData {
	if arg == nil {
		return nil
	}

	return &ProvinceResponseData{
		ID:   arg.ID,
		Name: arg.Name,
	}
}

func SerializeProvinces(arg []*entity.Province) []*ProvinceResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*ProvinceResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeProvince(arg[i]))
	}

	return res
}
