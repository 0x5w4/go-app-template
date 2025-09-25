package serializer

import (
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service"
	"time"
)

type SupportFeatureResponseData struct {
	ID        uint   `json:"id"`
	Code      string `json:"code,omitempty"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func SerializeSupportFeature(arg *entity.SupportFeature) *SupportFeatureResponseData {
	if arg == nil {
		return nil
	}

	return &SupportFeatureResponseData{
		ID:        arg.ID,
		Code:      arg.Code,
		Name:      arg.Name,
		Key:       arg.Key,
		IsActive:  arg.IsActive,
		CreatedAt: arg.CreatedAt.Format(time.RFC3339),
		UpdatedAt: arg.UpdatedAt.Format(time.RFC3339),
	}
}

func SerializeSupportFeatures(arg []*entity.SupportFeature) []*SupportFeatureResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*SupportFeatureResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeSupportFeature(arg[i]))
	}

	return res
}

type ValidatableStringResponseData struct {
	Value   string `json:"value"`
	Message string `json:"message,omitempty"`
}

type ValidatableBoolResponseData struct {
	Value   *bool  `json:"value"`
	Message string `json:"message,omitempty"`
}

type SupportFeaturePreviewResponseData struct {
	Row      int                           `json:"row"`
	Name     ValidatableStringResponseData `json:"name"`
	Key      ValidatableStringResponseData `json:"key"`
	IsActive ValidatableBoolResponseData   `json:"is_active"`
}

func SerializeSupportFeaturePreview(arg *service.SupportFeaturePreview) *SupportFeaturePreviewResponseData {
	if arg == nil {
		return nil
	}

	res := &SupportFeaturePreviewResponseData{
		Row: arg.Row,
		Name: ValidatableStringResponseData{
			Value:   arg.Name.Value,
			Message: arg.Name.Message,
		},
		Key: ValidatableStringResponseData{
			Value:   arg.Key.Value,
			Message: arg.Key.Message,
		},
		IsActive: ValidatableBoolResponseData{
			Value:   arg.IsActive.Value,
			Message: arg.IsActive.Message,
		},
	}

	return res
}

func SerializeSupportFeaturePreviews(arg []*service.SupportFeaturePreview) []*SupportFeaturePreviewResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*SupportFeaturePreviewResponseData, 0, len(arg))

	for _, item := range arg {
		if item == nil {
			continue
		}

		res = append(res, SerializeSupportFeaturePreview(item))
	}

	return res
}
