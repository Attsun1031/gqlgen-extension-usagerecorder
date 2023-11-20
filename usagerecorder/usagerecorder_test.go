package usagerecorder

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestGraphqlUsageRecorder_CollectGraphqlUsage(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		inputOperationContext  *graphql.OperationContext
		complexity             *extension.ComplexityStats
		expectedReferenceTypes map[string][]*ReferenceType
		expectedServiceName    string
	}{
		"valid": {
			expectedServiceName: "test",
			complexity:          &extension.ComplexityStats{Complexity: 10, ComplexityLimit: 30},
			inputOperationContext: &graphql.OperationContext{
				RawQuery:      "byId {\nid\nname\n}",
				Variables:     map[string]interface{}{"id": 1},
				OperationName: "GetSymptom",
				Operation: &ast.OperationDefinition{
					Name: "Query",
					SelectionSet: []ast.Selection{
						&ast.Field{
							Name: "byId",
							SelectionSet: []ast.Selection{
								&ast.Field{Name: "id"},
								&ast.Field{Name: "name"},
							},
						},
					},
				},
				Stats: graphql.Stats{OperationStart: now},
			},
			expectedReferenceTypes: map[string][]*ReferenceType{
				"byId": {
					{
						TypeName: "byId",
						Fields:   []string{"id", "name"},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// setup
			if tc.complexity != nil {
				tc.inputOperationContext.Stats.SetExtension("ComplexityLimit", tc.complexity)
			}
			ctx := context.Background()

			// Execute
			expectedMilliseconds := int64(100)
			expectedExtraValues := map[string]interface{}{"source_service": "source"}
			end := tc.inputOperationContext.Stats.OperationStart.Add(time.Millisecond * time.Duration(expectedMilliseconds))
			sut := New(
				WithClock(func() time.Time { return end }),
				WithExternalValuesExtractor(func(ctx context.Context, oc *graphql.OperationContext) map[string]interface{} {
					return expectedExtraValues
				}),
				WithEmitVariables(false),
			)
			got := sut.CollectGraphqlUsage(ctx, tc.inputOperationContext)

			// Verify
			var v map[string]interface{}
			assert.Equal(t, now, got.OperationTime)
			assert.Equal(t, tc.inputOperationContext.RawQuery, got.Query)
			assert.Equal(t, v, got.QueryVariables)
			assert.Equal(t, *tc.complexity, got.QueryComplexity)
			assert.Equal(t, tc.expectedReferenceTypes, got.ReferencedTypes)
			assert.Equal(t, expectedMilliseconds, got.OperationMilliseconds)
			assert.Equal(t, expectedExtraValues, got.ExtraValues)
		})
	}
}

func TestGraphqlUsageRecorder_GetUserAgent(t *testing.T) {
	cases := map[string]struct {
		inputUserAgent []string
		expected       string
	}{
		"has value": {
			inputUserAgent: []string{"test"},
			expected:       "test",
		},
		"no value": {
			inputUserAgent: []string{},
			expected:       "unknown",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			inputOperationContext := &graphql.OperationContext{
				Headers: map[string][]string{
					"User-Agent": tc.inputUserAgent,
				},
			}
			sut := New()
			got := sut.GetUserAgent(inputOperationContext)
			assert.Equal(t, tc.expected, got)
		})
	}
}
