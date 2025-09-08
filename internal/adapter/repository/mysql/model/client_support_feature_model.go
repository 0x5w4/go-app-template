package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type ClientSupportFeature struct {
	bun.BaseModel    `bun:"table:client_support_features,alias:clisf"`
	ClientID         uint            `bun:"client_id,pk"`
	Client           *Client         `bun:"rel:belongs-to,join:client_id=id"`
	SupportFeatureID uint            `bun:"support_feature_id,pk"`
	SupportFeature   *SupportFeature `bun:"rel:belongs-to,join:support_feature_id=id"`
	Order            int             `bun:"order,notnull"`
}

func (m *ClientSupportFeature) ToDomain() *entity.ClientSupportFeature {
	if m == nil {
		return nil
	}

	res := &entity.ClientSupportFeature{
		ClientID:         m.ClientID,
		SupportFeatureID: m.SupportFeatureID,
		Order:            m.Order,
	}
	if m.SupportFeature != nil {
		res.SupportFeature = m.SupportFeature.ToDomain()
	}

	return res
}

func ToClientSupportFeaturesDomain(arg []*ClientSupportFeature) []*entity.ClientSupportFeature {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.ClientSupportFeature, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsClientSupportFeature(arg *entity.ClientSupportFeature) *ClientSupportFeature {
	if arg == nil {
		return nil
	}

	return &ClientSupportFeature{
		ClientID:         arg.ClientID,
		Client:           AsClient(arg.Client),
		SupportFeatureID: arg.SupportFeatureID,
		SupportFeature:   AsSupportFeature(arg.SupportFeature),
		Order:            arg.Order,
	}
}

func AsClientSupportFeatures(arg []*entity.ClientSupportFeature) []*ClientSupportFeature {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*ClientSupportFeature, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsClientSupportFeature(arg[i]))
	}

	return res
}
