package model

import (
	"goapptemp/internal/domain/entity"
	"time"

	"github.com/uptrace/bun"
)

type Client struct {
	bun.BaseModel `bun:"table:clients,alias:cli"`
	Base
	CompanyID             uint                    `bun:"company_id,notnull"`
	Company               *Company                `bun:"rel:belongs-to,join:company_id=id"`
	DistrictID            uint                    `bun:"district_id,notnull"`
	District              *District               `bun:"rel:belongs-to,join:district_id=id"`
	ClientSupportFeatures []*ClientSupportFeature `bun:"rel:has-many,join:id=client_id"`
	Code                  string                  `bun:"code,notnull"`
	Name                  string                  `bun:"name,notnull"`
	Phone                 string                  `bun:"phone,notnull"`
	Fax                   *string                 `bun:"fax"`
	Icon                  *string                 `bun:"icon"`
	IconUpdatedAt         *time.Time              `bun:"icon_updated_at"`
	PICName               string                  `bun:"pic_name,notnull"`
	PICPhone              string                  `bun:"pic_phone,notnull"`
	Village               string                  `bun:"village,notnull"`
	PostalCode            string                  `bun:"postal_code,notnull"`
	Address               string                  `bun:"address,notnull"`
	CodeActive            *string                 `bun:"code_active,scanonly"`
	NameActive            *string                 `bun:"name_active,scanonly"`
}

func (m *Client) ToDomain() *entity.Client {
	if m == nil {
		return nil
	}

	res := &entity.Client{
		CompanyID:     m.CompanyID,
		DistrictID:    m.DistrictID,
		Code:          m.Code,
		Name:          m.Name,
		Phone:         m.Phone,
		Fax:           m.Fax,
		Icon:          m.Icon,
		IconUpdatedAt: m.IconUpdatedAt,
		PICName:       m.PICName,
		PICPhone:      m.PICPhone,
		Village:       m.Village,
		PostalCode:    m.PostalCode,
		Address:       m.Address,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
	if m.Company != nil {
		res.Company = m.Company.ToDomain()
	}

	if m.District != nil {
		res.District = m.District.ToDomain()
	}

	if len(m.ClientSupportFeatures) > 0 {
		res.ClientSupportFeatures = ToClientSupportFeaturesDomain(m.ClientSupportFeatures)
	}

	return res
}

func ToClientsDomain(arg []*Client) []*entity.Client {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.Client, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsClient(arg *entity.Client) *Client {
	if arg == nil {
		return nil
	}

	return &Client{
		CompanyID:             arg.CompanyID,
		Company:               AsCompany(arg.Company),
		DistrictID:            arg.DistrictID,
		District:              AsDistrict(arg.District),
		Code:                  arg.Code,
		Name:                  arg.Name,
		Phone:                 arg.Phone,
		Fax:                   arg.Fax,
		Icon:                  arg.Icon,
		IconUpdatedAt:         arg.IconUpdatedAt,
		PICName:               arg.PICName,
		PICPhone:              arg.PICPhone,
		Village:               arg.Village,
		PostalCode:            arg.PostalCode,
		Address:               arg.Address,
		ClientSupportFeatures: AsClientSupportFeatures(arg.ClientSupportFeatures),
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsClients(arg []*entity.Client) []*Client {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*Client, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsClient(arg[i]))
	}

	return res
}
