package hook

import (
	"context"

	"github.com/uptrace/bun"

	apm "go.elastic.co/apm/v2"
)

var _ bun.QueryHook = (*TracerHook)(nil)

type TracerHook struct{}

func NewTracerHook() *TracerHook {
	return &TracerHook{}
}

func (h *TracerHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	span, ctx := apm.StartSpan(ctx, "SQL "+event.Operation(), "db.query")
	ctx = apm.ContextWithSpan(ctx, span)

	return ctx
}

func (h *TracerHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
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
