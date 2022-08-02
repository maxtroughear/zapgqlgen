package zapgqlgen

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

type zapGqlgenContextKey struct{}

type ZapExtension struct {
	Logger      *zap.Logger
	UseNewRelic bool
}

var _ interface {
	graphql.HandlerExtension
	graphql.FieldInterceptor
} = ZapExtension{}

func (n ZapExtension) ExtensionName() string {
	return "ZapExtension"
}

func (n ZapExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (n ZapExtension) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	oc := graphql.GetOperationContext(ctx)
	fc := graphql.GetFieldContext(ctx)

	loggerFields := []zap.Field{
		zap.String("operation", oc.OperationName),
		zap.String("field", fc.Field.Name),
	}

	logger := n.Logger.With(loggerFields...)

	if n.UseNewRelic {
		nr := newrelic.FromContext(ctx)
		metadata := nr.GetLinkingMetadata()
		nrLoggerFields := []zap.Field{
			zap.String("entity.name", metadata.EntityName),
			zap.String("entity.guid", metadata.EntityGUID),
			zap.String("entity.type", metadata.EntityType),
			zap.String("hostname", metadata.Hostname),
			zap.String("trace.id", metadata.TraceID),
			zap.String("span.id", metadata.SpanID),
		}
		logger = logger.With(nrLoggerFields...)
	}

	ctx = newCtx(ctx, logger)
	return next(ctx)
}

func newCtx(ctx context.Context, ctxLogger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapGqlgenContextKey{}, ctxLogger)
}

func FromContext(ctx context.Context) *zap.Logger {
	l, _ := ctx.Value(zapGqlgenContextKey{}).(*zap.Logger)
	return l
}
