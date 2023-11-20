// Package usagerecorder provides a graphql extension to record graphql usage.
// This extension records graphql queries and variables from any service.
package usagerecorder

import (
	"context"
	"log/slog"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/Attsun1031/gqlgen-extension-usagerecorder/usagerecorder/model"
)

// GraphqlUsageRecorder records usages of graphql API calls. That is used to investigate API usages.
type GraphqlUsageRecorder struct {
	emitter              UsageEmitter
	clock                func() time.Time
	extraValuesExtractor func(ctx context.Context, oc *graphql.OperationContext) map[string]interface{}
	emitterErrorHandler  func(err error)
	logger               *slog.Logger
	emitVariables        bool
}

type UsageRecorderOption func(*GraphqlUsageRecorder)

// WithEmitter sets the emitter to the recorder.
func WithEmitter(emitter UsageEmitter) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.emitter = emitter
	}
}

// WithClock sets the clock to the recorder.
func WithClock(clock func() time.Time) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.clock = clock
	}
}

// WithExternalValuesExtractor sets the external values extractor to the recorder.
func WithExternalValuesExtractor(extractor func(ctx context.Context, oc *graphql.OperationContext) map[string]interface{}) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.extraValuesExtractor = extractor
	}
}

// WithEmitterErrorHandler sets the emitter error handler to the recorder.
func WithEmitterErrorHandler(handler func(err error)) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.emitterErrorHandler = handler
	}
}

// WithLogger sets the logger to the recorder.
func WithLogger(logger *slog.Logger) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.logger = logger
	}
}

// WithEmitVariables sets the emitVariables flag to the recorder.
func WithEmitVariables(emitVariables bool) UsageRecorderOption {
	return func(r *GraphqlUsageRecorder) {
		r.emitVariables = emitVariables
	}
}

// New creates a new GraphqlUsageRecorder.
func New(options ...UsageRecorderOption) *GraphqlUsageRecorder {
	defaultLogger := slog.Default()
	recorder := &GraphqlUsageRecorder{
		emitter: &LogEmitter{Logger: defaultLogger},
		clock:   time.Now,
		logger:  defaultLogger,
		extraValuesExtractor: func(ctx context.Context, oc *graphql.OperationContext) map[string]interface{} {
			return make(map[string]interface{})
		},
		emitterErrorHandler: func(err error) {
			defaultLogger.Error("failed to emit graphql usage", err)
		},
		emitVariables: false,
	}

	for _, opt := range options {
		opt(recorder)
	}

	return recorder
}

// Guarantee GraphqlUsageRecorder implements interfaces.
// See https://go.dev/doc/faq#guarantee_satisfies_interface
var _ interface {
	graphql.HandlerExtension
	graphql.ResponseInterceptor
} = &GraphqlUsageRecorder{}

const extensionName = "usage-recorder"

func (g *GraphqlUsageRecorder) ExtensionName() string {
	return extensionName
}

func (g *GraphqlUsageRecorder) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (g *GraphqlUsageRecorder) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	if oc.Operation == nil {
		// This block is for invalid query
		ua := g.GetUserAgent(oc)
		g.logger.Warn("operation is nil.", "userAgent", ua)
		return next(ctx)
	}
	if oc.Operation.Name == "IntrospectionQuery" {
		return next(ctx)
	}

	defer func() {
		usage := g.CollectGraphqlUsage(ctx, oc)
		if err := g.emitter.Emit(usage); err != nil {
			g.emitterErrorHandler(err)
		}
	}()
	return next(ctx)
}

func (g *GraphqlUsageRecorder) CollectGraphqlUsage(ctx context.Context, oc *graphql.OperationContext) *model.GraphqlUsage {
	complexityStats, ok := oc.Stats.GetExtension("ComplexityLimit").(*extension.ComplexityStats)
	if !ok {
		g.logger.Warn("failed to cast ComplexityStats")
		complexityStats = &extension.ComplexityStats{
			Complexity:      0,
			ComplexityLimit: 0,
		}
	}

	// extract dependent objects and fields
	cfs := graphql.CollectFields(oc, oc.Operation.SelectionSet, nil)
	var queryToReferencedTypes = make(map[string][]*model.ReferenceType, len(cfs))
	for _, cf := range cfs {
		refTypes := extractReferenceTypes(oc, cf, make([]*model.ReferenceType, 0))
		queryToReferencedTypes[cf.Name] = refTypes
	}
	end := g.clock()
	duration := end.Sub(oc.Stats.OperationStart)

	usage := &model.GraphqlUsage{
		OperationTime:         oc.Stats.OperationStart,
		QueryOperationName:    oc.Operation.Name,
		QueryComplexity:       *complexityStats,
		Query:                 oc.RawQuery,
		ReferencedTypes:       queryToReferencedTypes,
		OperationMilliseconds: duration.Milliseconds(),
		ExtraValues:           g.extraValuesExtractor(ctx, oc),
	}
	if g.emitVariables {
		usage.QueryVariables = oc.Variables
	}
	return usage
}

func (g *GraphqlUsageRecorder) GetUserAgent(oc *graphql.OperationContext) string {
	ua, ok := oc.Headers["User-Agent"]
	if !ok || len(ua) == 0 {
		return "unknown"
	}
	return ua[0]
}

func extractReferenceTypes(oc *graphql.OperationContext, cf graphql.CollectedField, result []*model.ReferenceType) []*model.ReferenceType {
	var fieldNames = make([]string, 0)
	for _, f := range graphql.CollectFields(oc, cf.Selections, nil) {
		if f.Selections != nil && len(f.Selections) > 0 {
			result = extractReferenceTypes(oc, f, nil)
		}
		if f.Name != "" {
			fieldNames = append(fieldNames, f.Name)
		}
	}
	result = append(result, &model.ReferenceType{TypeName: cf.Name, Fields: fieldNames})
	return result
}
