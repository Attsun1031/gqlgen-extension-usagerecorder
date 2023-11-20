package usagerecorder

import (
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"
)

type ReferenceType struct {
	TypeName string   `json:"typeName"`
	Fields   []string `json:"fields"`
}

type GraphqlUsage struct {
	OperationTime         time.Time                   `json:"requestedTimestamp"`
	QueryOperationName    string                      `json:"queryOperationName"`
	QueryComplexity       extension.ComplexityStats   `json:"queryComplexity"`
	QueryVariables        map[string]interface{}      `json:"queryVariables"`
	Query                 string                      `json:"query"`
	ReferencedTypes       map[string][]*ReferenceType `json:"referencedTypes"`
	OperationMilliseconds int64                       `json:"operationMilliseconds"`
	ExtraValues           map[string]interface{}      `json:"extraValues"`
}
