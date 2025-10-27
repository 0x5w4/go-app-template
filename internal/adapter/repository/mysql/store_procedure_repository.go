package mysqlrepository

import (
	"context"
	"database/sql"
	"fmt"
	"goapptemp/constant"
	"goapptemp/pkg/logger"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/uptrace/bun"
)

var _ StoreProcedureRepository = (*storeProcedureRepository)(nil)

type StoreProcedureRepository interface {
	CheckIfRecordsAreDeletable(ctx context.Context, parentTableName string, parentRecordIDs []uint, ignoreTables string) (map[uint]int, error)
}

type storeProcedureRepository struct {
	dbName string
	db     bun.IDB
	logger logger.Logger
}

func NewStoreProcedureRepository(dbName string, db bun.IDB, logger logger.Logger) *storeProcedureRepository {
	return &storeProcedureRepository{dbName: dbName, db: db, logger: logger}
}

func (r *storeProcedureRepository) CheckIfRecordsAreDeletable(ctx context.Context, parentTableName string, parentRecordIDs []uint, ignoreTables string) (map[uint]int, error) {
	var parentPK string

	err := r.db.NewSelect().
		Column("column_name").
		Table("information_schema.columns").
		Where("table_schema = ?", constant.ParentSchema).
		Where("table_name = ?", parentTableName).
		Where("column_key = 'PRI'").
		Limit(1).
		Scan(ctx, &parentPK)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get parent PK for table '%s': primary key not found", parentTableName)
		}

		return nil, fmt.Errorf("failed to get parent PK for table '%s': %w", parentTableName, err)
	}

	if parentPK == "" {
		return nil, fmt.Errorf("failed to get parent PK for table '%s': primary key not found", parentTableName)
	}

	type FKReference struct {
		ChildSchema   string       `bun:"child_schema"`
		ChildTable    string       `bun:"child_table"`
		ChildFKColumn string       `bun:"child_fk_column"`
		HasSoftDelete sql.NullBool `bun:"has_soft_delete_column"`
	}

	var refs []FKReference

	query := `
     SELECT DISTINCT
       kcu.table_schema AS child_schema,
       kcu.table_name AS child_table,
       kcu.column_name AS child_fk_column,
       EXISTS (
         SELECT 1
         FROM information_schema.columns AS c
         WHERE c.table_schema = kcu.table_schema
         AND c.table_name = kcu.table_name
         AND c.column_name = ?
       ) AS has_soft_delete_column
     FROM information_schema.key_column_usage AS kcu
     WHERE kcu.referenced_table_schema = ?
       AND kcu.referenced_table_name = ?
       AND kcu.referenced_column_name = ?
   `

	if err := r.db.NewRaw(query, constant.SoftDeleteColumnName, constant.ParentSchema, parentTableName, parentPK).
		Scan(ctx, &refs); err != nil {
		return nil, fmt.Errorf("error fetching foreign key references: %w", err)
	}

	if len(refs) == 0 {
		return make(map[uint]int), nil
	}

	idStrings := make([]string, 0, len(parentRecordIDs))
	for _, id := range parentRecordIDs {
		idStrings = append(idStrings, strconv.FormatUint(uint64(id), 10))
	}

	idList := strings.Join(idStrings, ",")
	if idList == "" {
		return make(map[uint]int), nil
	}

	unionQueries := make([]string, 0)

	for _, ref := range refs {
		if ignoreTables != "" && strings.Contains(ignoreTables, ref.ChildTable) {
			continue
		}

		q := fmt.Sprintf(`
          SELECT %s AS parent_id, COUNT(*) AS dependency_count
          FROM %s.%s
          WHERE %s IN (%s)`,
			ref.ChildFKColumn,
			ref.ChildSchema,
			ref.ChildTable,
			ref.ChildFKColumn,
			idList,
		)

		if ref.HasSoftDelete.Valid && ref.HasSoftDelete.Bool {
			q += fmt.Sprintf(" AND %s IS NULL", constant.SoftDeleteColumnName)
		}

		q += " GROUP BY " + ref.ChildFKColumn
		unionQueries = append(unionQueries, q)
	}

	if len(unionQueries) == 0 {
		return make(map[uint]int), nil
	}

	fullQuery := "SELECT parent_id, SUM(dependency_count) AS total_dependencies FROM (" +
		strings.Join(unionQueries, " UNION ALL ") + ") AS final_counts GROUP BY parent_id"

	type DependencyResult struct {
		ParentID          uint `bun:"parent_id"`
		TotalDependencies int  `bun:"total_dependencies"`
	}

	var scanResults []DependencyResult
	if err := r.db.NewRaw(fullQuery).Scan(ctx, &scanResults); err != nil {
		return nil, fmt.Errorf("executing dependency check failed: %w", err)
	}

	result := make(map[uint]int)
	for _, item := range scanResults {
		result[item.ParentID] = item.TotalDependencies
	}

	return result, nil
}
