package apm

import (
	"context"

	"github.com/uptrace/bun"

	"go.elastic.co/apm/v2"
)

var _ bun.QueryHook = (*QueryHook)(nil)

type QueryHook struct{}

func NewQueryHook() *QueryHook {
	return &QueryHook{}
}

func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	span, ctx := apm.StartSpan(ctx, "SQL "+event.Operation(), "db.query")
	ctx = apm.ContextWithSpan(ctx, span)

	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	span := apm.SpanFromContext(ctx)
	if span != nil {
		truncatedQuery := event.Query
		if len(truncatedQuery) > 100 {
			truncatedQuery = truncatedQuery[:100] + "..."
		}

		span.Context.SetLabel("query", truncatedQuery)
		span.End()
	}
}
