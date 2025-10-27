package service

import (
	"bytes"
	"context"
	"fmt"
	"goapptemp/config"
	"goapptemp/constant"
	"goapptemp/internal/adapter/repository"
	mysqlrepository "goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	serror "goapptemp/internal/domain/service/error"

	"github.com/cockroachdb/errors"
	validator "github.com/go-playground/validator/v10"
	excelize "github.com/xuri/excelize/v2"
)

var _ SupportFeatureService = (*supportFeatureService)(nil)

type SupportFeatureService interface {
	Create(ctx context.Context, req *CreateSupportFeatureRequest) (*entity.SupportFeature, error)
	BulkCreate(ctx context.Context, req *BulkCreateSupportFeatureRequest) ([]*entity.SupportFeature, error)
	Update(ctx context.Context, req *UpdateSupportFeatureRequest) (*entity.SupportFeature, error)
	Delete(ctx context.Context, req *DeleteSupportFeatureRequest) error
	Find(ctx context.Context, req *FindSupportFeaturesRequest) ([]*entity.SupportFeature, int, error)
	FindOne(ctx context.Context, req *FindOneSupportFeatureRequest) (*entity.SupportFeature, error)
	IsDeletable(ctx context.Context, req *IsDeletableSupportFeatureRequest) (bool, error)
	ImportPreview(ctx context.Context, req *ImportPreviewSupportFeatureRequest) ([]*SupportFeaturePreview, error)
	TemplateImport(ctx context.Context, req *TemplateImportSupportFeatureRequest) (*FileServiceData, error)
}

type supportFeatureService struct {
	config   *config.Config
	repo     repository.Repository
	logger   logger.Logger
	auth     AuthService
	validate *validator.Validate
}

func NewSupportFeatureService(config *config.Config, repo repository.Repository, logger logger.Logger, auth AuthService, validate *validator.Validate) *supportFeatureService {
	return &supportFeatureService{
		config:   config,
		repo:     repo,
		logger:   logger,
		auth:     auth,
		validate: validate,
	}
}

type FileServiceData struct {
	Filename string
	MIMEType string
	Content  io.Reader
	Size     int64
}

type CreateSupportFeatureRequest struct {
	AuthParams     *AuthParams
	SupportFeature *entity.SupportFeature
}

func (s *supportFeatureService) Create(ctx context.Context, req *CreateSupportFeatureRequest) (*entity.SupportFeature, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.CREATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.SupportFeature == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Data cannot be nil")
	}

	var supportFeature *entity.SupportFeature

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		for {
			sfCode, err := shared.GenerateCode(constant.CodePefix["support_feature"], 5)
			if err != nil {
				return err
			}

			exists, err := s.repo.MySQL().SupportFeature().CheckCodeExists(ctx, sfCode)
			if err != nil {
				return err
			}

			if !exists {
				req.SupportFeature.Code = sfCode
				break
			}
		}

		var err error

		supportFeature, err = s.repo.MySQL().SupportFeature().Create(ctx, req.SupportFeature)
		if err != nil {
			return err
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return supportFeature, nil
}

type BulkCreateSupportFeatureRequest struct {
	AuthParams      *AuthParams
	SupportFeatures []*entity.SupportFeature
}

func (s *supportFeatureService) BulkCreate(ctx context.Context, req *BulkCreateSupportFeatureRequest) ([]*entity.SupportFeature, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.CREATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if len(req.SupportFeatures) == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Data cannot be nil")
	}

	if len(req.SupportFeatures) > 300 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "bulk create supports up to 300 at a time")
	}

	var supportFeaturesToReturn []*entity.SupportFeature

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		finalCodesForItems := make([]string, len(req.SupportFeatures))
		assignedCodesSet := make(map[string]struct{})
		initialPhaseCodes := make([]string, len(req.SupportFeatures))
		codeToOriginalIndexMap := make(map[string]int)

		for i := range req.SupportFeatures {
			var candidateCode string

			for range 100 {
				generatedCand, err := shared.GenerateCode(constant.CodePefix["support_feature"], 5)
				if err != nil {
					return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to generate code")
				}

				if _, exists := assignedCodesSet[generatedCand]; !exists {
					candidateCode = generatedCand
					break
				}
			}

			if candidateCode == "" {
				return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to generate unique code within batch")
			}

			initialPhaseCodes[i] = candidateCode
			assignedCodesSet[candidateCode] = struct{}{}
			codeToOriginalIndexMap[candidateCode] = i
		}

		transactionalSFRepo := txRepo.SupportFeature()

		dbExistingCodes, err := transactionalSFRepo.GetExistingCodes(ctx, initialPhaseCodes)
		if err != nil {
			return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to check existing codes in db")
		}

		itemsToRegenerateIndices := []int{}

		for code, originalIndex := range codeToOriginalIndexMap {
			if _, existsInDB := dbExistingCodes[code]; existsInDB {
				itemsToRegenerateIndices = append(itemsToRegenerateIndices, originalIndex)

				delete(assignedCodesSet, code)
			} else {
				finalCodesForItems[originalIndex] = code
			}
		}

		if len(itemsToRegenerateIndices) > 0 {
			for _, itemIndex := range itemsToRegenerateIndices {
				var newCodeForItem string

				for range 100 {
					candidateCode, err := shared.GenerateCode(constant.CodePefix["support_feature"], 5)
					if err != nil {
						return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to generate code")
					}

					if _, existsInBatch := assignedCodesSet[candidateCode]; existsInBatch {
						continue
					}

					existsInDB, err := transactionalSFRepo.CheckCodeExists(ctx, candidateCode)
					if err != nil {
						return exception.Wrap(err, exception.TypeInternalError, exception.CodeInternalError, "Failed to check code existence in db")
					}

					if !existsInDB {
						newCodeForItem = candidateCode
						break
					}
				}

				if newCodeForItem == "" {
					return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to generate code after all attempts")
				}

				finalCodesForItems[itemIndex] = newCodeForItem
				assignedCodesSet[newCodeForItem] = struct{}{}
			}
		}

		for i := range req.SupportFeatures {
			if finalCodesForItems[i] == "" {
				return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Code not generated")
			}

			req.SupportFeatures[i].Code = finalCodesForItems[i]
		}

		createdFeatures, bulkCreateErr := transactionalSFRepo.BulkCreate(ctx, req.SupportFeatures)
		if bulkCreateErr != nil {
			return bulkCreateErr
		}

		supportFeaturesToReturn = createdFeatures

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return supportFeaturesToReturn, nil
}

type FindSupportFeaturesRequest struct {
	AuthParams *AuthParams
	Filter     *mysqlrepository.FilterSupportFeaturePayload
}

func (s *supportFeatureService) Find(ctx context.Context, req *FindSupportFeaturesRequest) ([]*entity.SupportFeature, int, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, 0, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.READ")
	if err != nil {
		return nil, 0, err
	}

	if !ok {
		return nil, 0, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	supportFeatures, totalCount, err := s.repo.MySQL().SupportFeature().Find(ctx, req.Filter)
	if err != nil {
		return nil, 0, serror.TranslateRepoError(err)
	}

	return supportFeatures, totalCount, nil
}

type FindOneSupportFeatureRequest struct {
	AuthParams       *AuthParams
	SupportFeatureID uint
}

func (s *supportFeatureService) FindOne(ctx context.Context, req *FindOneSupportFeatureRequest) (*entity.SupportFeature, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.READ")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.SupportFeatureID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Help service ID required for find one")
	}

	supportFeature, err := s.repo.MySQL().SupportFeature().FindByID(ctx, req.SupportFeatureID)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return supportFeature, nil
}

type UpdateSupportFeatureRequest struct {
	AuthParams *AuthParams
	Update     *mysqlrepository.UpdateSupportFeaturePayload
}

func (s *supportFeatureService) Update(ctx context.Context, req *UpdateSupportFeatureRequest) (*entity.SupportFeature, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.UPDATE")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.Update == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Update payload cannot be nil")
	}

	if req.Update.ID == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Help service ID required for update")
	}

	supportFeature, err := s.repo.MySQL().SupportFeature().Update(ctx, req.Update)
	if err != nil {
		return nil, serror.TranslateRepoError(err)
	}

	return supportFeature, nil
}

type DeleteSupportFeatureRequest struct {
	AuthParams       *AuthParams
	SupportFeatureID uint
}

func (s *supportFeatureService) Delete(ctx context.Context, req *DeleteSupportFeatureRequest) error {
	if req.AuthParams.AccessTokenClaims == nil {
		return exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.DELETE")
	if err != nil {
		return err
	}

	if !ok {
		return exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.SupportFeatureID == 0 {
		return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Help service ID cannot be zero")
	}

	atomicOperation := func(txRepo mysqlrepository.MySQLRepository) error {
		sfTable := txRepo.SupportFeature().GetTableName()

		dependencyMap, err := txRepo.StoreProcedure().CheckIfRecordsAreDeletable(ctx, sfTable, []uint{req.SupportFeatureID}, "")
		if err != nil {
			return err
		}

		if count, found := dependencyMap[req.SupportFeatureID]; found && count > 0 {
			return exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Help service is not deletable due to existing dependencies")
		}

		if err := txRepo.SupportFeature().Delete(ctx, req.SupportFeatureID); err != nil {
			return err
		}

		return nil
	}
	if err := s.repo.MySQL().Atomic(ctx, s.config, atomicOperation); err != nil {
		return serror.TranslateRepoError(err)
	}

	return nil
}

type IsDeletableSupportFeatureRequest struct {
	AuthParams       *AuthParams
	SupportFeatureID uint
}

func (s *supportFeatureService) IsDeletable(ctx context.Context, req *IsDeletableSupportFeatureRequest) (bool, error) {
	if req.AuthParams.AccessTokenClaims == nil {
		return false, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, err := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.DELETE")
	if err != nil {
		return false, err
	}

	if !ok {
		return false, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.SupportFeatureID == 0 {
		return false, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Help service ID required for check deletable")
	}

	sfTable := s.repo.MySQL().SupportFeature().GetTableName()

	dependencyMap, err := s.repo.MySQL().StoreProcedure().CheckIfRecordsAreDeletable(ctx, sfTable, []uint{req.SupportFeatureID}, "")
	if err != nil {
		return false, serror.TranslateRepoError(err)
	}

	if count, found := dependencyMap[req.SupportFeatureID]; found && count > 0 {
		return false, nil
	}

	return true, nil
}

type TemplateImportSupportFeatureRequest struct {
	AuthParams *AuthParams
	File       *multipart.FileHeader
}

func (s *supportFeatureService) TemplateImport(ctx context.Context, req *TemplateImportSupportFeatureRequest) (*FileServiceData, error) {
	const (
		mainSheetName     = "Help Services"
		maxDataRows       = 300
		headerRow         = 1
		dataStartRow      = 2
		totalRowsToFormat = headerRow + maxDataRows
	)

	headers := []struct {
		Name         string
		ColumnLetter string
		CommentText  string
		Width        float64
	}{
		{Name: "Name", ColumnLetter: "A", CommentText: "Required. Alphabets and spaces only, 2 to 32 characters.", Width: 40},
		{Name: "Key", ColumnLetter: "B", CommentText: "Required. Alphabets, underscore (_) only. No spaces/numbers. 2 to 32 characters.", Width: 45},
		{Name: "Is Active", ColumnLetter: "C", CommentText: "Required. Select TRUE or FALSE.", Width: 15},
	}

	f := excelize.NewFile()

	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to close excel file")
		}
	}()

	mainSheetIndex, err := f.NewSheet(mainSheetName)
	if err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create main sheet")
	}

	f.SetActiveSheet(mainSheetIndex)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
		},
		Protection: &excelize.Protection{Locked: true},
	})
	if err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create header style")
	}

	for i := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		if err = f.SetCellValue(mainSheetName, cell, headers[i].Name); err != nil {
			return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set cell value for %s: %v", cell, err))
		}

		if err = f.SetCellStyle(mainSheetName, cell, cell, headerStyle); err != nil {
			return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set cell style for %s: %v", cell, err))
		}

		if err = f.SetColWidth(mainSheetName, headers[i].ColumnLetter, headers[i].ColumnLetter, headers[i].Width); err != nil {
			return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set column width for %s: %v", headers[i].ColumnLetter, err))
		}

		comment := excelize.Comment{
			Cell:   cell,
			Author: "Template Guide:",
			Paragraph: []excelize.RichTextRun{
				{Text: "Guidance: ", Font: &excelize.Font{Bold: true, Color: "000000"}},
				{Text: headers[i].CommentText, Font: &excelize.Font{Color: "000000"}},
			},
			Height: 70,
			Width:  300,
		}
		if err = f.AddComment(mainSheetName, comment); err != nil {
			return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to add comment for %s: %v", cell, err))
		}
	}

	if err = f.SetRowHeight(mainSheetName, headerRow, 30); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set row height for row %d: %v", headerRow, err))
	}

	dvName := excelize.NewDataValidation(true)
	dvName.Sqref = fmt.Sprintf("A%d:A%d", dataStartRow, totalRowsToFormat)

	if err = dvName.SetRange(float64(2), float64(32), excelize.DataValidationTypeTextLength, excelize.DataValidationOperatorBetween); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set range for Name data validation: %v", err))
	}

	errTitleName := "Invalid Name"
	dvName.ErrorTitle = &errTitleName
	errMsgName := "Name must be 2 to 32 characters, containing only alphabets and spaces."
	dvName.Error = &errMsgName
	dvName.ShowErrorMessage = true
	promptTitleName := "Name Input"
	dvName.PromptTitle = &promptTitleName
	promptName := "Enter a name (alphabets and spaces, 2-32 characters)."
	dvName.Prompt = &promptName
	dvName.ShowInputMessage = true

	if err = f.AddDataValidation(mainSheetName, dvName); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to add Name data validation")
	}

	dvKey := excelize.NewDataValidation(true)
	dvKey.Sqref = fmt.Sprintf("B%d:B%d", dataStartRow, totalRowsToFormat)

	if err = dvKey.SetRange(float64(2), float64(32), excelize.DataValidationTypeTextLength, excelize.DataValidationOperatorBetween); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set range for Key data validation: %v", err))
	}

	errTitleKey := "Invalid Key"
	dvKey.ErrorTitle = &errTitleKey
	errMsgKey := "Key: 2-32 chars (letters, _). No spaces/numbers."
	dvKey.Error = &errMsgKey
	dvKey.ShowErrorMessage = true
	promptTitleKey := "Key Input"
	dvKey.PromptTitle = &promptTitleKey
	promptKey := "Enter key (2-32 chars: letters, _). No spaces/numbers. E.g., 'feature_key'."
	dvKey.Prompt = &promptKey
	dvKey.ShowInputMessage = true

	if err = f.AddDataValidation(mainSheetName, dvKey); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to add Key data validation")
	}

	dvIsActive := excelize.NewDataValidation(true)
	dvIsActive.Sqref = fmt.Sprintf("C%d:C%d", dataStartRow, totalRowsToFormat)

	if err := dvIsActive.SetDropList([]string{"TRUE", "FALSE"}); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to set dropdown list for IsActive")
	}

	errTitleIsActive := "Invalid Value"
	dvIsActive.ErrorTitle = &errTitleIsActive
	errMsgIsActive := "Please select TRUE or FALSE from the list."
	dvIsActive.Error = &errMsgIsActive
	dvIsActive.ShowErrorMessage = true
	promptTitleIsActive := "Activation Status"
	dvIsActive.PromptTitle = &promptTitleIsActive
	promptIsActive := "Select if the feature is active."
	dvIsActive.Prompt = &promptIsActive
	dvIsActive.ShowInputMessage = true

	if err := f.AddDataValidation(mainSheetName, dvIsActive); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to add IsActive data validation")
	}

	unlockedStyle, err := f.NewStyle(&excelize.Style{
		Protection: &excelize.Protection{Locked: false},
		Alignment:  &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create unlocked style")
	}

	for rNum := dataStartRow; rNum <= totalRowsToFormat; rNum++ {
		for cNum := 1; cNum <= len(headers); cNum++ {
			cell, _ := excelize.CoordinatesToCellName(cNum, rNum)
			if err = f.SetCellStyle(mainSheetName, cell, cell, unlockedStyle); err != nil {
				return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, fmt.Sprintf("Failed to set cell style for %s: %v", cell, err))
			}
		}
	}

	sheetProtectionOptions := &excelize.SheetProtectionOptions{
		Password:            "",
		SelectLockedCells:   true,
		SelectUnlockedCells: true,
		EditObjects:         false, EditScenarios: false, FormatCells: false, FormatColumns: false,
		FormatRows: false, InsertColumns: false, InsertRows: false, InsertHyperlinks: false,
		DeleteColumns: false, DeleteRows: false, Sort: false, AutoFilter: false, PivotTables: false,
	}
	if err := f.ProtectSheet(mainSheetName, sheetProtectionOptions); err != nil {
		return nil, exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to protect main sheet")
	}

	if err := createErrorGuideSheet(f); err != nil {
		return nil, err
	}

	defaultSheetName := "Sheet1"
	if mainSheetName != defaultSheetName {
		if sheetIdx, _ := f.GetSheetIndex(defaultSheetName); sheetIdx != -1 {
			if defaultSheetName != "Error Guide" {
				f.DeleteSheet(defaultSheetName)
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	contentBytes := buf.Bytes()

	return &FileServiceData{
		Filename: "help_service_import_template.xlsx",
		MIMEType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Content:  bytes.NewReader(contentBytes),
		Size:     int64(len(contentBytes)),
	}, nil
}

func createErrorGuideSheet(f *excelize.File) error {
	const errorSheetName = "Error Guide"

	_, err := f.NewSheet(errorSheetName)
	if err != nil {
		return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create error guide sheet")
	}

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "000000"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create title style")
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
		},
	})
	if err != nil {
		return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create header style")
	}

	cellStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
		},
	})
	if err != nil {
		return exception.New(exception.TypeInternalError, exception.CodeInternalError, "Failed to create cell style")
	}

	f.SetColWidth(errorSheetName, "A", "A", 40)
	f.SetColWidth(errorSheetName, "B", "B", 60)
	f.SetCellValue(errorSheetName, "A1", "Invalid data format")
	f.SetCellStyle(errorSheetName, "A1", "A1", titleStyle)
	f.MergeCell(errorSheetName, "A1", "B1")
	f.SetCellValue(errorSheetName, "A3", "Error Description")
	f.SetCellValue(errorSheetName, "B3", "Common Case / Example")
	f.SetCellStyle(errorSheetName, "A3", "B3", headerStyle)
	f.SetRowHeight(errorSheetName, 3, 30)

	invalidDataFormatErrors := [][]string{
		{"Data cannot be empty", "'Name' column is left blank."},
		{"Invalid data format", "'Date' column is '30 May 2025' instead of a letter or numeric format (if applicable, example based on image context)."},
		{"Data length invalid", "'Phone Number' column filled with more than 15 digits."},
		{"Value not a valid type", "'Transaction Nominal' column filled with text 'One hundred thousand rupiahs', should be numbers (e.g., 100000)."},
		{"Value not found", "'Region Code' column filled with 'XYZ123' which does not exist in the reference system."},
		{"Duplicate in file", "Rows 15 and 17 have the same value 'KEY123' in the Key column."},
	}
	currentRow := 4

	for _, data := range invalidDataFormatErrors {
		f.SetCellValue(errorSheetName, fmt.Sprintf("A%d", currentRow), data[0])
		f.SetCellValue(errorSheetName, fmt.Sprintf("B%d", currentRow), data[1])
		f.SetCellStyle(errorSheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("B%d", currentRow), cellStyle)
		f.SetRowHeight(errorSheetName, currentRow, 45)

		currentRow++
	}

	currentRow += 1
	title2Cell := fmt.Sprintf("A%d", currentRow)
	f.SetCellValue(errorSheetName, title2Cell, "Data already exists")
	f.SetCellStyle(errorSheetName, title2Cell, title2Cell, titleStyle)
	f.MergeCell(errorSheetName, title2Cell, fmt.Sprintf("B%d", currentRow))

	currentRow++
	header2CellA := fmt.Sprintf("A%d", currentRow)
	header2CellB := fmt.Sprintf("B%d", currentRow)

	f.SetCellValue(errorSheetName, header2CellA, "Error Description")
	f.SetCellValue(errorSheetName, header2CellB, "Common Case / Example")
	f.SetCellStyle(errorSheetName, header2CellA, header2CellB, headerStyle)
	f.SetRowHeight(errorSheetName, currentRow, 30)

	currentRow++

	dataAlreadyExistsErrors := [][]string{
		{"Data already available in the system", "'Key' column contains 'KEY123' which has been previously added."},
		{"Duplicate with existing data", "'Client' data already exists with active status."},
	}
	for _, data := range dataAlreadyExistsErrors {
		f.SetCellValue(errorSheetName, fmt.Sprintf("A%d", currentRow), data[0])
		f.SetCellValue(errorSheetName, fmt.Sprintf("B%d", currentRow), data[1])
		f.SetCellStyle(errorSheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("B%d", currentRow), cellStyle)
		f.SetRowHeight(errorSheetName, currentRow, 45)

		currentRow++
	}

	return nil
}

type ImportPreviewSupportFeatureRequest struct {
	AuthParams *AuthParams
	File       *multipart.FileHeader
}

type ValidatableString struct {
	Value   string `json:"value"             validate:"required,min=2,max=50,alpha_space"`
	Message string `json:"message,omitempty"`
}

type ValidatableKey struct {
	Value   string `json:"value"             validate:"required,min=2,max=50,username_chars_allowed"`
	Message string `json:"message,omitempty"`
}

type ValidatableBool struct {
	Value   *bool  `json:"value"             validate:"required,boolean"`
	Message string `json:"message,omitempty"`
}

type SupportFeaturePreview struct {
	Row      int               `json:"row"`
	Name     ValidatableString `json:"name"      validate:"required"`
	Key      ValidatableKey    `json:"key"       validate:"required"`
	IsActive ValidatableBool   `json:"is_active" validate:"required"`
}

func (s *supportFeatureService) ImportPreview(ctx context.Context, req *ImportPreviewSupportFeatureRequest) ([]*SupportFeaturePreview, error) {
	if req.AuthParams == nil || req.AuthParams.AccessTokenClaims == nil {
		return nil, exception.New(exception.TypePermissionDenied, exception.CodeForbidden, "Token payload not provided")
	}

	ok, authErr := s.auth.AuthorizationCheck(ctx, req.AuthParams.AccessTokenClaims.UserID, "HELP_SERVICE.CREATE")
	if authErr != nil {
		return nil, exception.Wrap(authErr, exception.TypeInternalError, exception.CodeInternalError, "Authorization check failed")
	}

	if !ok {
		return nil, exception.New(exception.TypeForbidden, exception.CodeForbidden, "Not allowed to access")
	}

	if req.File == nil {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Missing data. Please fill in the required field.")
	}

	src, fileOpenErr := req.File.Open()
	if fileOpenErr != nil {
		return nil, exception.Wrap(fileOpenErr, exception.TypeInternalError, exception.CodeInternalError, "Failed to open uploaded file")
	}

	defer func() {
		if err := src.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to close uploaded file")
		}
	}()

	f, excelOpenErr := excelize.OpenReader(src)
	if excelOpenErr != nil {
		return nil, exception.Wrap(excelOpenErr, exception.TypeBadRequest, exception.CodeBadRequest, "Invalid data format. Please check your entry.")
	}

	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Error().Err(err).Msg("Failed to close excel file")
		}
	}()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Invalid data format. Please check your entry.")
	}

	sheetName := sheetList[0]

	rows, rowsErr := f.GetRows(sheetName)
	if rowsErr != nil {
		return nil, exception.Wrap(rowsErr, exception.TypeInternalError, exception.CodeInternalError, "Failed to read rows from excel sheet: "+sheetName)
	}

	if len(rows) < 2 {
		return make([]*SupportFeaturePreview, 0), nil
	}

	header := rows[0]
	colMap := make(map[string]int)

	for i, colName := range header {
		colMap[strings.TrimSpace(colName)] = i
	}

	requiredCols := []string{"Name", "Key", "Is Active"}

	var missingCols []string

	maxRequiredIndexProcessed := -1

	for _, reqCol := range requiredCols {
		idx, exists := colMap[reqCol]
		if !exists {
			missingCols = append(missingCols, reqCol)
		} else if idx > maxRequiredIndexProcessed {
			maxRequiredIndexProcessed = idx
		}
	}

	if len(missingCols) > 0 {
		return nil, exception.New(exception.TypeBadRequest, exception.CodeBadRequest, "Missing column(s) in header: "+strings.Join(missingCols, ", ")+". Please fill in the required field.")
	}

	previews := make([]*SupportFeaturePreview, 0, len(rows)-1)
	keysToValidate := make([]string, 0, len(rows)-1)
	namesToValidate := make([]string, 0, len(rows)-1)

	for i, row := range rows[1:] {
		excelRowNumber := i + 2
		sf := &SupportFeaturePreview{Row: excelRowNumber}
		nameColIdx := colMap["Name"]

		if len(row) > nameColIdx {
			sf.Name.Value = strings.TrimSpace(row[nameColIdx])
		}

		keyColIdx := colMap["Key"]
		if len(row) > keyColIdx {
			sf.Key.Value = strings.TrimSpace(row[keyColIdx])
		}

		isActiveColIdx := colMap["Is Active"]
		if len(row) > isActiveColIdx {
			isActiveStr := strings.TrimSpace(strings.ToLower(row[isActiveColIdx]))
			if isActiveStr != "" {
				if parsedBool, err := strconv.ParseBool(isActiveStr); err == nil {
					sf.IsActive.Value = &parsedBool
				}
			}
		}

		previews = append(previews, sf)

		if sf.Name.Value != "" {
			namesToValidate = append(namesToValidate, sf.Name.Value)
		}

		if sf.Key.Value != "" {
			keysToValidate = append(keysToValidate, sf.Key.Value)
		}
	}

	existingKeysInDB, existingNamesInDB, dbErr := s.repo.MySQL().SupportFeature().FindExistingKeysAndNames(ctx, keysToValidate, namesToValidate)
	if dbErr != nil {
		return nil, serror.TranslateRepoError(dbErr)
	}

	allProcessedSFs := make([]*SupportFeaturePreview, 0, len(previews))
	errSFs := make([]*SupportFeaturePreview, 0, len(previews))
	seenKeysInFile := make(map[string]int)
	seenNamesInFile := make(map[string]int)

	for _, sf := range previews {
		if err := s.validate.Struct(sf); err != nil {
			var validationErrors validator.ValidationErrors
			if errors.As(err, &validationErrors) {
				for _, fe := range validationErrors {
					var message string

					switch fe.Tag() {
					case "required":
						if fe.StructNamespace() == "SupportFeaturePreview.IsActive.Value" {
							message = "Missing data. Please select in the required field."
						} else {
							message = "Missing data. Please fill in the required field."
						}
					default:
						message = "Invalid data format. Please check your entry."
					}

					switch fe.StructNamespace() {
					case "SupportFeaturePreview.Name.Value":
						if sf.Name.Message == "" {
							sf.Name.Message = message
						}
					case "SupportFeaturePreview.Key.Value":
						if sf.Key.Message == "" {
							sf.Key.Message = message
						}
					case "SupportFeaturePreview.IsActive.Value":
						if sf.IsActive.Message == "" {
							sf.IsActive.Message = message
						}
					}
				}
			}
		}

		if sf.Name.Message == "" && sf.Name.Value != "" {
			lowerName := strings.ToLower(sf.Name.Value)
			if _, nameAlreadySeen := seenNamesInFile[lowerName]; nameAlreadySeen {
				sf.Name.Message = "Duplicate entry found in file. Each entry in the file must be unique."
			} else {
				seenNamesInFile[lowerName] = sf.Row

				if _, nameExists := existingNamesInDB[lowerName]; nameExists {
					sf.Name.Message = "Data already exists. Please enter new information."
				}
			}
		}

		if sf.Key.Message == "" && sf.Key.Value != "" {
			lowerKey := strings.ToLower(sf.Key.Value)
			if _, keyAlreadySeen := seenKeysInFile[lowerKey]; keyAlreadySeen {
				sf.Key.Message = "Duplicate entry found in file. Each entry in the file must be unique."
			} else {
				seenKeysInFile[lowerKey] = sf.Row

				if _, keyExists := existingKeysInDB[lowerKey]; keyExists {
					sf.Key.Message = "Data already exists. Please enter new information."
				}
			}
		}

		if sf.Name.Message != "" || sf.Key.Message != "" || sf.IsActive.Message != "" {
			errSFs = append(errSFs, sf)
		} else {
			allProcessedSFs = append(allProcessedSFs, sf)
		}
	}

	if len(errSFs) > 0 {
		return errSFs, nil
	}

	return allProcessedSFs, nil
}
