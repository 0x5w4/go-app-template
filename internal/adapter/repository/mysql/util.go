package mysqlrepository

import "github.com/uptrace/bun"

func applyMultiLikeFilter(q *bun.SelectQuery, fieldExpr string, values []string) *bun.SelectQuery {
	if len(values) == 0 {
		return q
	}

	return q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
		for i := range values {
			sq = sq.WhereOr("LOWER("+fieldExpr+") LIKE LOWER(?)", "%"+values[i]+"%")
		}

		return sq
	})
}
