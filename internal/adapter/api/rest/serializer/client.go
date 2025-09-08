package serializer

import (
	"goapptemp/internal/domain/entity"
	"sort"
	"time"
)

type ClientSupportFeatureResponseData struct {
	ID       uint   `json:"id"`
	Code     string `json:"code,omitempty"`
	Name     string `json:"name,omitempty"`
	Key      string `json:"key,omitempty"`
	IsActive bool   `json:"is_active,omitempty"`
	Order    int    `json:"order"`
}

type ClientResponseData struct {
	ID              uint                                `json:"id"`
	CompanyID       uint                                `json:"company_id,omitempty"`
	Company         *CompanyResponseData                `json:"company,omitempty"`
	Code            string                              `json:"code,omitempty"`
	Name            string                              `json:"name,omitempty"`
	Phone           string                              `json:"phone,omitempty"`
	Fax             *string                             `json:"fax,omitempty"`
	Icon            *string                             `json:"icon,omitempty"`
	PICName         string                              `json:"pic_name,omitempty"`
	PICPhone        string                              `json:"pic_phone,omitempty"`
	DistrictID      uint                                `json:"district_id,omitempty"`
	District        *DistrictResponseData               `json:"district,omitempty"`
	Village         string                              `json:"village,omitempty"`
	PostalCode      string                              `json:"postal_code,omitempty"`
	Address         string                              `json:"address,omitempty"`
	SupportFeatures []*ClientSupportFeatureResponseData `json:"help_services,omitempty"`
	CreatedAt       string                              `json:"created_at,omitempty"`
	UpdatedAt       string                              `json:"updated_at,omitempty"`
}

func SerializeClient(arg *entity.Client) *ClientResponseData {
	if arg == nil {
		return nil
	}

	res := &ClientResponseData{
		ID:         arg.ID,
		CompanyID:  arg.CompanyID,
		Company:    SerializeCompany(arg.Company),
		Code:       arg.Code,
		Name:       arg.Name,
		Phone:      arg.Phone,
		Fax:        arg.Fax,
		Icon:       arg.Icon,
		PICName:    arg.PICName,
		PICPhone:   arg.PICPhone,
		DistrictID: arg.DistrictID,
		District:   SerializeDistrict(arg.District),
		Village:    arg.Village,
		PostalCode: arg.PostalCode,
		Address:    arg.Address,
		CreatedAt:  arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  arg.UpdatedAt.Format(time.RFC3339),
	}

	if arg.ClientSupportFeatures != nil {
		res.SupportFeatures = make([]*ClientSupportFeatureResponseData, 0, len(arg.ClientSupportFeatures))

		for _, csf := range arg.ClientSupportFeatures {
			if csf != nil && csf.SupportFeature != nil {
				supportFeatureResponse := &ClientSupportFeatureResponseData{
					ID:    csf.SupportFeatureID,
					Code:  csf.SupportFeature.Code,
					Key:   csf.SupportFeature.Key,
					Name:  csf.SupportFeature.Name,
					Order: csf.Order,
				}
				res.SupportFeatures = append(res.SupportFeatures, supportFeatureResponse)
			}
		}

		sort.Slice(res.SupportFeatures, func(i, j int) bool {
			return res.SupportFeatures[i].Order < res.SupportFeatures[j].Order
		})
	}

	return res
}

func SerializeClients(arg []*entity.Client) []*ClientResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*ClientResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeClient(arg[i]))
	}

	return res
}
