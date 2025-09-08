package supportfeature

import (
	"io"
	"mime/multipart"

	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/auth"
)

type FileServiceData struct {
	Filename string
	MIMEType string
	Content  io.Reader
	Size     int64
}

type CreateSupportFeatureRequest struct {
	AuthParams     *auth.AuthParams
	SupportFeature *entity.SupportFeature
}

type BulkCreateSupportFeatureRequest struct {
	AuthParams      *auth.AuthParams
	SupportFeatures []*entity.SupportFeature
}

type UpdateSupportFeatureRequest struct {
	AuthParams *auth.AuthParams
	Update     *mysql.UpdateSupportFeaturePayload
}

type DeleteSupportFeatureRequest struct {
	AuthParams       *auth.AuthParams
	SupportFeatureID uint
}

type FindSupportFeaturesRequest struct {
	AuthParams *auth.AuthParams
	Filter     *mysql.FilterSupportFeaturePayload
}

type FindOneSupportFeatureRequest struct {
	AuthParams       *auth.AuthParams
	SupportFeatureID uint
}

type IsDeletableSupportFeatureRequest struct {
	AuthParams       *auth.AuthParams
	SupportFeatureID uint
}

type ImportPreviewSupportFeatureRequest struct {
	AuthParams *auth.AuthParams
	File       *multipart.FileHeader
}

type TemplateImportSupportFeatureRequest struct {
	AuthParams *auth.AuthParams
	File       *multipart.FileHeader
}
