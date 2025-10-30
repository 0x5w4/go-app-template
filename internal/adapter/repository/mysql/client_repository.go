package mysqlrepository

import (
	"context"
	"fmt"
	"goapptemp/internal/adapter/repository/mysql/model"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/shared/exception"
	"goapptemp/pkg/logger"
	"time"

	"github.com/uptrace/bun"
)

var _ ClientRepository = (*clientRepository)(nil)

type ClientRepository interface {
	GetTableName() string
	Create(ctx context.Context, req *entity.Client) (*entity.Client, error)
	FindByID(ctx context.Context, id uint, isWithRelation bool) (*entity.Client, error)
	Find(ctx context.Context, filter *FilterClientPayload) ([]*entity.Client, int, error)
	Update(ctx context.Context, req *UpdateClientPayload) (*entity.Client, error)
	Delete(ctx context.Context, id uint) error
	UpdateStaleIcons(ctx context.Context) error
	IsCodeExists(ctx context.Context, code string) (bool, error)
}

type clientRepository struct {
	db     bun.IDB
	logger logger.Logger
}

func NewClientRepository(db bun.IDB, logger logger.Logger) *clientRepository {
	return &clientRepository{db: db, logger: logger}
}

func (r *clientRepository) GetTableName() string {
	return "clients"
}

func (r *clientRepository) Create(ctx context.Context, req *entity.Client) (*entity.Client, error) {
	if req == nil {
		return nil, handleDBError(exception.ErrDataNull, r.GetTableName(), "create client")
	}

	client := model.AsClient(req)
	if _, err := r.db.NewInsert().Model(client).Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "create client")
	}

	return client.ToDomain(), nil
}

type FilterClientPayload struct {
	IDs        []uint
	CompanyIDs []uint
	Codes      []string
	Names      []string
	PICNames   []string
	Search     string
	Page       int
	PerPage    int
}

func (r *clientRepository) Find(ctx context.Context, filter *FilterClientPayload) ([]*entity.Client, int, error) {
	var clients []*model.Client

	query := r.db.NewSelect().Model(&clients).
		Relation("District.City.Province").
		Relation("Company")
	if len(filter.IDs) > 0 {
		query = query.Where("cli.id IN (?)", bun.In(filter.IDs))
	}

	if len(filter.CompanyIDs) > 0 {
		query = query.Where("cli.company_id IN (?)", bun.In(filter.CompanyIDs))
	}

	query = applyMultiLikeFilter(query, "cli.code", filter.Codes)
	query = applyMultiLikeFilter(query, "cli.name", filter.Names)
	query = applyMultiLikeFilter(query, "cli.pic_name", filter.PICNames)

	if filter.Search != "" {
		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.WhereOr("LOWER(cli.code) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(cli.name) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(cli.pic_name) LIKE LOWER(?)", "%"+filter.Search+"%")
			q = q.WhereOr("LOWER(cli.pic_phone) LIKE LOWER(?)", "%"+filter.Search+"%")

			return q
		})
	}

	totalCount, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "count client")
	}

	if totalCount == 0 {
		return []*entity.Client{}, 0, nil
	}

	if filter.PerPage > 0 {
		query = query.Limit(filter.PerPage)
	}

	if filter.Page > 0 && filter.PerPage > 0 {
		offset := (filter.Page - 1) * filter.PerPage
		query = query.Offset(offset)
	}

	query = query.Order("cli.id DESC")
	if err := query.Scan(ctx); err != nil {
		return nil, 0, handleDBError(err, r.GetTableName(), "find client")
	}

	return model.ToClientsDomain(clients), totalCount, nil
}

func (r *clientRepository) FindByID(ctx context.Context, id uint, isWithRelation bool) (*entity.Client, error) {
	if id == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "find client by id")
	}

	client := &model.Client{Base: model.Base{ID: id}}

	query := r.db.NewSelect().Model(client).WherePK()
	if isWithRelation {
		query = query.
			Relation("District.City.Province").
			Relation("Company").
			Relation("ClientSupportFeatures.SupportFeature")
	}

	if err := query.Scan(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), fmt.Sprintf("find client by id: %d", id))
	}

	return client.ToDomain(), nil
}

type UpdateClientPayload struct {
	ID                    uint
	CompanyID             *uint
	Name                  *string
	ProfileID             *uint
	Phone                 *string
	Fax                   *string
	Icon                  *string
	IconUpdatedAt         *time.Time
	PICName               *string
	PICPhone              *string
	DistrictID            *uint
	Village               *string
	PostalCode            *string
	Address               *string
	ClientSupportFeatures []*entity.ClientSupportFeature
}

func (r *clientRepository) Update(ctx context.Context, req *UpdateClientPayload) (*entity.Client, error) {
	if req.ID == 0 {
		return nil, handleDBError(exception.ErrIDNull, r.GetTableName(), "update client")
	}

	client := &model.Client{Base: model.Base{ID: req.ID}}

	var columnsToUpdate []string

	if req.CompanyID != nil {
		client.CompanyID = *req.CompanyID

		columnsToUpdate = append(columnsToUpdate, "company_id")
	}

	if req.Name != nil {
		client.Name = *req.Name

		columnsToUpdate = append(columnsToUpdate, "name")
	}

	if req.Phone != nil {
		client.Phone = *req.Phone

		columnsToUpdate = append(columnsToUpdate, "phone")
	}

	if req.Fax != nil {
		client.Fax = req.Fax

		columnsToUpdate = append(columnsToUpdate, "fax")
	}

	if req.Icon != nil {
		client.Icon = req.Icon

		columnsToUpdate = append(columnsToUpdate, "icon")
	}

	if req.IconUpdatedAt != nil {
		client.IconUpdatedAt = req.IconUpdatedAt

		columnsToUpdate = append(columnsToUpdate, "icon_updated_at")
	}

	if req.PICName != nil {
		client.PICName = *req.PICName

		columnsToUpdate = append(columnsToUpdate, "pic_name")
	}

	if req.PICPhone != nil {
		client.PICPhone = *req.PICPhone

		columnsToUpdate = append(columnsToUpdate, "pic_phone")
	}

	if req.DistrictID != nil {
		client.DistrictID = *req.DistrictID

		columnsToUpdate = append(columnsToUpdate, "district_id")
	}

	if req.Village != nil {
		client.Village = *req.Village

		columnsToUpdate = append(columnsToUpdate, "village")
	}

	if req.PostalCode != nil {
		client.PostalCode = *req.PostalCode

		columnsToUpdate = append(columnsToUpdate, "postal_code")
	}

	if req.Address != nil {
		client.Address = *req.Address

		columnsToUpdate = append(columnsToUpdate, "address")
	}

	if len(columnsToUpdate) == 0 {
		return client.ToDomain(), nil
	}

	query := r.db.NewUpdate().Model(client).Column(columnsToUpdate...).WherePK()
	if _, err := query.Returning("*").Exec(ctx); err != nil {
		return nil, handleDBError(err, r.GetTableName(), "update client")
	}

	return client.ToDomain(), nil
}

func (r *clientRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return handleDBError(exception.ErrIDNull, r.GetTableName(), "delete client")
	}

	client := &model.Client{Base: model.Base{ID: id}}
	if _, err := r.db.NewDelete().Model(client).WherePK().Exec(ctx); err != nil {
		return handleDBError(err, r.GetTableName(), "delete client")
	}

	return nil
}

func (r *clientRepository) UpdateStaleIcons(ctx context.Context) error {
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)

	query := r.db.NewUpdate().
		Model((*model.Client)(nil)).
		Set("icon = ?", "failed").
		Where("icon = ?", "loading").
		Where("icon_updated_at < ?", thirtySecondsAgo)
	if _, err := query.Exec(ctx); err != nil {
		return handleDBError(err, r.GetTableName(), "update stale icons")
	}

	return nil
}

func (r *clientRepository) IsCodeExists(ctx context.Context, code string) (bool, error) {
	if code == "" {
		return false, handleDBError(exception.ErrDataNull, r.GetTableName(), "check client code exists")
	}

	exist, err := r.db.NewSelect().
		Model((*model.Client)(nil)).
		Where("LOWER(code_active) = LOWER(?)", code).
		Exists(ctx)
	if err != nil {
		return false, handleDBError(err, r.GetTableName(), "check client code exists")
	}

	return exist, nil
}
