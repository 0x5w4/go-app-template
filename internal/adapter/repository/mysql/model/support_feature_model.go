package model

import (
	"goapptemp/internal/domain/entity"

	"github.com/uptrace/bun"
)

type SupportFeature struct {
	bun.BaseModel `bun:"table:support_features,alias:sft"`
	Base
	Code       string  `bun:"code,notnull,unique"`
	Name       string  `bun:"name,notnull"`
	Key        string  `bun:"key,notnull,unique"`
	IsActive   bool    `bun:"is_active,notnull"`
	CodeActive *string `bun:"code_active,scanonly"`
	NameActive *string `bun:"name_active,scanonly"`
	KeyActive  *string `bun:"key_active,scanonly"`
}

func (m *SupportFeature) ToDomain() *entity.SupportFeature {
	if m == nil {
		return nil
	}

	return &entity.SupportFeature{
		Code:     m.Code,
		Name:     m.Name,
		Key:      m.Key,
		IsActive: m.IsActive,
		Base: entity.Base{
			ID:        m.ID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
}

func ToSupportFeaturesDomain(arg []*SupportFeature) []*entity.SupportFeature {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*entity.SupportFeature, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, arg[i].ToDomain())
	}

	return res
}

func AsSupportFeature(arg *entity.SupportFeature) *SupportFeature {
	if arg == nil {
		return nil
	}

	return &SupportFeature{
		Code:     arg.Code,
		Name:     arg.Name,
		Key:      arg.Key,
		IsActive: arg.IsActive,
		Base: Base{
			ID:        arg.ID,
			CreatedAt: arg.CreatedAt,
			UpdatedAt: arg.UpdatedAt,
			DeletedAt: arg.DeletedAt,
		},
	}
}

func AsSupportFeatures(arg []*entity.SupportFeature) []*SupportFeature {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*SupportFeature, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, AsSupportFeature(arg[i]))
	}

	return res
}
